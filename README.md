# Go-event-scheduling

## Objective (Given)

In a geographically distributed team, it is very hard to find common time to meet that works for everyone. Your objective is to build an API, using which we will solve this problem for any event.

The organizer creates an event with a brief title “Brainstorming meeting” and provides. N slots (eg 12 Jan 2025, 2 - 4PM EST, 14 Jan 2025 6-9 PM EST etc.) also provide estimated time required for the meeting eg. 1 hr.

All participants also provide their availability in the similar format. 

The system recommends the time slots that work for all. If there is no such time slot found, then it recommends time slots that work for the most number of people (also provides a list for whom it does not work).

## Core Functionality

    Implement the REST API in Golang.

    Support creating, updating, and deleting events

    Support creating, updating, and deleting preferred time slots by each user.

    Endpoint that shows the possible time slots for the event.

## Prerequisites

Before getting started, make sure you have the following installed:

- [Docker](https://www.docker.com/get-started)
- [Terraform](https://www.terraform.io/downloads.html)
- [Go](https://golang.org/dl/)
- [Nginx](https://nginx.org/en/docs/)

## Setup

### 1. Clone the Repository

Running the code on to your local machine:
    
    git clone https://github.com/Deepaks86/Go-event-scheduling.git
    cd Go-event-scheduling

## Go API Server for Event Sceduling
Run the server:

go run main.go

The Go server (located in server/) is a REST API for event scheduling. It uses several endpoints:

    POST /event - Create a new event
    GET /events/{id} - Get event details by ID
    PUT /event/{id} - Update an existing event
    DELETE /event/{id} - Delete an event
    POST /participant - Create a participant's availability
    GET /participant/{participant_id} - Get a participant's availability
    PUT /participant/{participant_id} - Update a participant's availability
    DELETE /participant/{participant_id}/event/{event_id} - Delete a participant's availability for an event
    GET /event/{id}/find-common-slots - Find common available slots for an event

To check API data you can use JSON requests given in "JSONrequests sample.docx"

## Running Automated Tests

go test -v
To run tests for the Go server, use the following command:

go test -v

This will run the unit tests defined in the main_test.go file.

## Containerization of the application
Dockerfile
    Go Server Dockerfile: Located in the server/ directory, this file defines how the Go server is built and run in a Docker container.
    Nginx Dockerfile: Located in the nginx/ directory, this file defines the Nginx container build process.

### Dockerfile Deployment

cd docker

        To deploy the project: 
        1. Build the Docker images:
            docker-compose build

        2. Deploy using Docker Compose:
            docker-compose up -d    
        This will start the containers in detached mode and make the API available via the configured port in docker-compose.yml.
        The nginx container will serve on port 8080 and loadbalances the requests to the Go containers

        To test create an event using URL: http://localhost:8080/event

        3. Shutdown: To stop the services:
            docker-compose down

## Design for horizontal scalability
In the docker-compose.yml file, the backend service is configured to scale horizontally using Docker's deploy settings. 

    Service Name: backend
    Scaling:
        The backend service is set to run 3 replicas (instances) to ensure that multiple backend containers are deployed.
        This configuration ensures that the backend can handle more requests by utilizing multiple instances of the backend service.
        To ensure that the same client IP is directed to the same backend server:
            ip_hash is configured in nginx.conf

## Infrastructure as Code (IaC) using Terraform
This project uses Terraform to manage infrastructure. You will need to set up your Terraform configuration before running it.
In the main.tf assuming GCP infrastructure the resources are deployed
    Note: Make sure you have configured your cloud provider credentials if required.
    
    Initialize Terraform:

terraform init

    Apply the Terraform configuration to provision resources:

terraform apply

Resources created: 
    1. event-scheduler-vm with
        metadata for installing docker, docker-compose, clone git repo and bringing up the containers using startup.sh
    2. firewall rule Allow-8080 over the internet
        To test create an event using URL: http://<Public IP of VM>:8080/event