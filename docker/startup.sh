#!/bin/bash
cd docker

# stop running containers if any
docker-compose down

# remove earlier images
docker image rm docker-nginx:latest
docker image rm docker-backend:latest

# remove all docker containers
echo "y" | sudo docker system prune -a

# build and run docker containers
docker-compose up -d
