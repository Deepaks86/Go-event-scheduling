package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *mux.Router {
	router := mux.NewRouter()

	// Event Routes
	router.HandleFunc("/event", createEvent).Methods("POST")
	router.HandleFunc("/events/{id}", getEvent).Methods("GET")
	router.HandleFunc("/event/{id}", updateEvent).Methods("PUT")
	router.HandleFunc("/event/{id}", deleteEvent).Methods("DELETE")

	// Participant Availability Routes
	router.HandleFunc("/participant", createParticipantAvailability).Methods("POST")
	router.HandleFunc("/participant/{participant_id}", getParticipantAvailability).Methods("GET")
	router.HandleFunc("/participant/{participant_id}", updateParticipantAvailability).Methods("PUT")
	router.HandleFunc("/participant/{participant_id}/event/{event_id}", deleteParticipantAvailability).Methods("DELETE")

	// Find common slots
	router.HandleFunc("/event/{id}/find-common-slots", findCommonSlots).Methods("GET")

	return router
}

func TestCreateEvent(t *testing.T) {
	router := setupRouter()

	event := Event{
		Title:         "Test Event",
		Slots:         []Slot{{StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour)}},
		EstimatedTime: 1 * time.Hour,
		Participants:  []string{"1", "2"},
	}

	eventJSON, _ := json.Marshal(event)
	req, err := http.NewRequest("POST", "/event", bytes.NewBuffer(eventJSON))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Expected status code 201")
}

func TestGetEvent(t *testing.T) {
	// Set up the router
	router := setupRouter()

	// Creating an event to test the GET endpoint
	event := Event{
		Title:         "Test Event",
		Slots:         []Slot{{StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour)}},
		EstimatedTime: 1 * time.Hour,
		Participants:  []string{"1", "2"},
	}

	eventID := "1"
	events[eventID] = event

	req, err := http.NewRequest("GET", "/events/"+eventID, nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

	// Check if event is in the response body
	var response Event
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	assert.Equal(t, event.Title, response.Title, "Event title should match")
}

func TestUpdateParticipantAvailability(t *testing.T) {
	// Set up the router
	router := setupRouter()

	// Creating a participant and event to test updating availability
	participant := Participant{
		ID:           "1",
		EventID:      "1",
		Availability: []Slot{{StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour)}},
	}
	participants[participant.ID] = append(participants[participant.ID], participant)

	updatedAvailability := []Slot{
		{StartTime: time.Now().Add(2 * time.Hour), EndTime: time.Now().Add(3 * time.Hour)},
	}

	// Prepare the request body
	updateRequest := struct {
		EventID string `json:"event_id"`
		Slots   []Slot `json:"slots"`
	}{
		EventID: "1",
		Slots:   updatedAvailability,
	}

	updateJSON, _ := json.Marshal(updateRequest)

	req, err := http.NewRequest("PUT", "/participant/1", bytes.NewBuffer(updateJSON))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

	// Check if participant's availability was updated
	var response map[string]string
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	assert.Equal(t, "Availability updated successfully", response["message"])
}

func TestFindCommonSlots(t *testing.T) {
	// Set up the router
	router := setupRouter()

	// Create an event and participants for this test
	event := Event{
		Title:         "Test Event",
		Slots:         []Slot{{StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour)}},
		EstimatedTime: 1 * time.Hour,
		Participants:  []string{"1", "2"},
	}
	events["1"] = event

	participant1 := Participant{
		ID:           "1",
		EventID:      "1",
		Availability: []Slot{{StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour)}},
	}
	participants["1"] = append(participants["1"], participant1)

	participant2 := Participant{
		ID:           "2",
		EventID:      "1",
		Availability: []Slot{{StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour)}},
	}
	participants["2"] = append(participants["2"], participant2)

	req, err := http.NewRequest("GET", "/event/1/find-common-slots", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

	// Check if the response contains recommended time slots
	var response AvailabilityResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	assert.NotEmpty(t, response.RecommendedTimeSlots, "There should be recommended time slots")
}

func TestDeleteParticipantAvailability(t *testing.T) {
	// Set up the router
	router := setupRouter()

	// Create an event first to ensure that the event exists before adding participants
	event := Event{
		ID:            "1",
		Title:         "Test Event",
		Slots:         []Slot{{StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour)}},
		EstimatedTime: 1 * time.Hour,
		Participants:  []string{}, // Initially, no participants
	}
	events[event.ID] = event

	// Create a participant and availability
	participant := Participant{
		ID:           "1",
		EventID:      "1",
		Availability: []Slot{{StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour)}},
	}

	// Add participant to the participants map
	participants["1"] = append(participants["1"], participant)

	// Add the participant to the event's participants list
	event = events["1"] // Fetch the event into a variable so we can modify it
	event.Participants = append(event.Participants, "1")
	events["1"] = event // Reassign the modified event back to the map

	// Prepare the request to delete the availability
	req, err := http.NewRequest("DELETE", "/participant/1/event/1", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

	// Check if the response contains a success message
	var response map[string]string
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	// Check if the participant's availability was deleted
	assert.Equal(t, "All slots for this event deleted successfully", response["message"])

	// Ensure that the participant has been removed from the events map
	if len(events["1"].Participants) > 0 {
		t.Fatalf("Expected participant to be removed from event")
	}
}

func TestDeleteEvent(t *testing.T) {
	// Set up the router
	router := setupRouter()

	// Create an event
	event := Event{
		Title:         "Test Event",
		Slots:         []Slot{{StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour)}},
		EstimatedTime: 1 * time.Hour,
		Participants:  []string{"1", "2"},
	}
	events["1"] = event

	// Prepare the request to delete the event
	req, err := http.NewRequest("DELETE", "/event/1", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code, "Expected status code 204")

	// Check if the event was deleted
	if _, exists := events["1"]; exists {
		t.Fatalf("Expected event to be deleted")
	}
}
