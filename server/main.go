package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Slot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type Event struct {
	ID            string        `json:"id"`
	Title         string        `json:"title"`
	Slots         []Slot        `json:"slots"`
	EstimatedTime time.Duration `json:"estimatedTime"`
	Participants  []string      `json:"participants"`
}

type Participant struct {
	ID           string `json:"id"`
	EventID      string `json:"event_id"`
	Availability []Slot `json:"availability"`
}

type ParticipantAvailability struct {
	Participant_ID string `json:"participant_id"`
	Slots          []Slot `json:"slots"`
}

type AvailabilityResponse struct {
	RecommendedTimeSlots []SlotUnavailable `json:"recommendedTimeSlots"`
}

type SlotUnavailable struct {
	Slot                    Slot     `json:"slot"`
	UnavailableParticipants []string `json:"unavailableParticipants"`
}

var events = make(map[string]Event)
var participants = make(map[string][]Participant)

// Create Event Handler
func createEvent(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the event details
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid input"})
		return
	}
	// Generate a unique ID for the event
	eventID := fmt.Sprintf("%d", len(events)+1)
	event.ID = eventID
	events[eventID] = event
	// Return a success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Event created successfully with ID: " + eventID})
}

// Get Event Handler
func getEvent(w http.ResponseWriter, r *http.Request) {
	// Extract event_id from the URL parameters
	params := mux.Vars(r)
	eventID := params["id"]
	_, exists := events[eventID]
	// If the event does not exist, return a 404 error
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Event not found"})
		return
	}
	// Return the event
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(events[eventID])
}

// Update Event Handler
func updateEvent(w http.ResponseWriter, r *http.Request) {
	// Extract event_id from the URL parameters
	params := mux.Vars(r)
	eventID := params["id"]
	// Parse the request body to get the updated event details
	var updatedEvent Event
	err := json.NewDecoder(r.Body).Decode(&updatedEvent)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid input"})
		return
	}
	event, exists := events[eventID]
	// If the event does not exist, return a 404 error
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Event not found"})
		return
	}
	// Update the event
	event.Title = updatedEvent.Title
	event.Slots = updatedEvent.Slots
	event.EstimatedTime = updatedEvent.EstimatedTime
	events[eventID] = event
	// Return a success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Event updated successfully"})
}

// Delete Event Handler
func deleteEvent(w http.ResponseWriter, r *http.Request) {
	// Extract event_id from the URL parameters
	params := mux.Vars(r)
	eventID := params["id"]
	_, exists := events[eventID]
	// If the event does not exist, return a 404 error
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Event not found"})
		return
	}
	// Delete the event
	delete(events, eventID)
	w.WriteHeader(http.StatusNoContent)
}

// Function to create the availability details of a participant for an event
func createParticipantAvailability(w http.ResponseWriter, r *http.Request) {
	// Define the struct to read the request body
	var availabilityRequest struct {
		Participant_ID string `json:"participant_id"`
		EventID        string `json:"event_id"`
		Slots          []Slot `json:"slots"`
	}
	// Parse the request body to get the event_id, participant_id, and availability slots
	err := json.NewDecoder(r.Body).Decode(&availabilityRequest)
	if err != nil {
		// If the input is invalid, return a 400 error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid input"})
		return
	}
	if _, exists := events[availabilityRequest.EventID]; !exists {
		// If the event does not exist, return a 404 error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Event not found"})
		return
	}
	for _, participant := range participants[availabilityRequest.Participant_ID] {
		// If the user is already associated with the provided event_id, return a 409 error
		if participant.EventID == availabilityRequest.EventID {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"message": "This availability has already been recorded"})
			return
		}
	}
	// Create a new Participant entry for this user and event
	participant := Participant{
		ID:           availabilityRequest.Participant_ID,
		EventID:      availabilityRequest.EventID,
		Availability: availabilityRequest.Slots,
	}
	// Add the participant to the event
	participants[availabilityRequest.Participant_ID] = append(participants[availabilityRequest.Participant_ID], participant)
	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Availability created successfully"})
}

// Function to get the availability details of a participant for an event
func getParticipantAvailability(w http.ResponseWriter, r *http.Request) {
	// Extract participant_id from the URL parameters
	params := mux.Vars(r)
	paricipantID := params["participant_id"]
	// Check if the participant exists in the participants map
	participant, exists := participants[paricipantID]
	if !exists {
		// If the participant does not exist, return a 404 error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Participant not found"})
		return
	}
	// Respond with the participant details and availability
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(participant)
}

// Function to update the availability details of a participant for an event
func updateParticipantAvailability(w http.ResponseWriter, r *http.Request) {
	// Extract participant_id from the URL parameters
	params := mux.Vars(r)
	paricipantID := params["participant_id"]
	// Define the struct to read the request body
	var availabilityRequest struct {
		EventID string `json:"event_id"`
		Slots   []Slot `json:"slots"`
	}

	// Parse the request body to get the new availability slots and event_id
	err := json.NewDecoder(r.Body).Decode(&availabilityRequest)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid input"})
		return
	}

	// Check if the event exists
	if _, exists := events[availabilityRequest.EventID]; !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Event not found"})
		return
	}
	// Check if the participant exists
	participantFound := false
	for i, participant := range participants[paricipantID] {
		// Check if the user is associated with the provided event_id
		if participant.EventID == availabilityRequest.EventID {
			// Update the availability slots for the participant
			participants[paricipantID][i].Availability = availabilityRequest.Slots
			participantFound = true
			break
		}
	}
	// If the participant is not found for the specified event_id, return a 404 error
	if !participantFound {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Participant or event not found"})
		return
	}
	// Respond with the updated participant availability
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Availability updated successfully"})
}

// Function to delete the availability details of a participant for an event
func deleteParticipantAvailability(w http.ResponseWriter, r *http.Request) {
	// Extract participant_id and event_id from the URL parameters
	params := mux.Vars(r)
	participantID := params["participant_id"]
	eventID := params["event_id"]

	event, exists := events[eventID]
	// Check if the event exists
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Event not found"})
		return
	}
	// Check if the user has availability for the event
	participantFound := false
	for i, participant := range participants[participantID] {
		if participant.EventID == eventID {
			// If the user is found for the event, remove all their slots for that event
			participantFound = true
			// Remove the participant from the participants map for that event
			participants[participantID] = append(participants[participantID][:i], participants[participantID][i+1:]...)
			// Also remove the participant from the event's Participants list (since event is a value, not a pointer)
			for j, pid := range event.Participants {
				if pid == participantID {
					// Remove the participant from the event's Participants list
					event.Participants = append(event.Participants[:j], event.Participants[j+1:]...)
					break
				}
			}

			// Reassign the updated event back to the map
			events[eventID] = event

			// Respond with success
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "All slots for this event deleted successfully"})
			return
		}
	}

	// If the user is not found for the event, return 404
	if !participantFound {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Participant or event not found"})
		return
	}
}

// Helper function to check if an event slot works for a user
func isSlotAvailableForUser(eventSlot Slot, participantAvailability ParticipantAvailability) bool {
	for _, userSlot := range participantAvailability.Slots {
		// Check if the user's slot overlaps with the event's slot
		if eventSlot.StartTime.Before(userSlot.EndTime) && eventSlot.EndTime.After(userSlot.StartTime) {
			// If there's an overlap (even partial), the user is considered available
			return true
		}
	}
	// No overlap, so the user is unavailable
	return false
}

// Find common slots for the event based on its participants' availability
func findCommonSlots(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	eventID := params["id"]

	event, exists := events[eventID]
	// Check if the event exists
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Event not found"})
		return
	}

	// Prepare the response structure
	var recommendedTimeSlots []SlotUnavailable
	var maxParticipants int
	var bestSlots []Slot

	// Process each slot for the current event (Event 1)
	for _, eventSlot := range event.Slots {
		var availableParticipants []string
		// Get participants for this specific event
		for _, paricipantID := range event.Participants {
			// Get the list of participants for this event
			var participantAvailability ParticipantAvailability
			for _, participant := range participants[paricipantID] {
				if participant.EventID == eventID {
					// Found the participant for this event
					participantAvailability = ParticipantAvailability{
						Participant_ID: paricipantID,
						Slots:          participant.Availability,
					}
					break
				}
			}

			// Check if user is available for the event slot
			if isSlotAvailableForUser(eventSlot, participantAvailability) {
				availableParticipants = append(availableParticipants, paricipantID)
			}
		}
		// If all participants are available, add the slot to recommended time slots
		if len(availableParticipants) == len(event.Participants) {
			recommendedTimeSlots = append(recommendedTimeSlots, SlotUnavailable{
				Slot:                    eventSlot,
				UnavailableParticipants: []string{}, // All participants are available
			})
		} else {
			// Track slots that work for the most participants
			if len(availableParticipants) > maxParticipants {
				maxParticipants = len(availableParticipants)
				bestSlots = []Slot{eventSlot}
			} else if len(availableParticipants) == maxParticipants {
				bestSlots = append(bestSlots, eventSlot)
			}
		}
	}

	// If no slots work for all, suggest the best slots for the most participants
	if len(recommendedTimeSlots) == 0 && len(bestSlots) > 0 {
		for _, eventSlot := range bestSlots {
			var unavailable []string
			// For each of the best slots, find unavailable participants
			for _, paricipantID := range event.Participants {
				var participantAvailability ParticipantAvailability
				// Get user availability for the current event
				for _, participant := range participants[paricipantID] {
					if participant.EventID == eventID {
						participantAvailability = ParticipantAvailability{
							Participant_ID: paricipantID,
							Slots:          participant.Availability,
						}
						break
					}
				}

				// If user is unavailable, add to the list of unavailable participants
				if !isSlotAvailableForUser(eventSlot, participantAvailability) {
					unavailable = append(unavailable, paricipantID)
				}
			}
			if unavailable == nil {
				unavailable = []string{}
			}
			// Add slot with unavailable participants
			recommendedTimeSlots = append(recommendedTimeSlots, SlotUnavailable{
				Slot:                    eventSlot,
				UnavailableParticipants: unavailable,
			})
		}
	}

	// Respond with the recommended time slots and unavailable participants
	response := AvailabilityResponse{
		RecommendedTimeSlots: recommendedTimeSlots,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	router := mux.NewRouter()

	// Event Routes
	router.HandleFunc("/event", createEvent).Methods("POST")
	router.HandleFunc("/events/{id}", getEvent).Methods("GET")
	router.HandleFunc("/event/{id}", updateEvent).Methods("PUT")
	router.HandleFunc("/event/{id}", deleteEvent).Methods("DELETE")

	// User Availability Routes
	router.HandleFunc("/participant", createParticipantAvailability).Methods("POST") // Get possible slots for the event
	router.HandleFunc("/participant/{participant_id}", getParticipantAvailability).Methods("GET")
	router.HandleFunc("/participant/{participant_id}", updateParticipantAvailability).Methods("PUT")
	router.HandleFunc("/participant/{participant_id}/event/{event_id}", deleteParticipantAvailability).Methods("DELETE")

	router.HandleFunc("/event/{id}/find-common-slots", findCommonSlots).Methods("GET")

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
