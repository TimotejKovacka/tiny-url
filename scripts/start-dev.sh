#!/bin/bash

NAMESPACE="tiny-url"
COMPOSE_FILE="./deployments/docker-compose.yaml"

echo "Starting Docker Compose for $NAMESPACE..."
docker-compose -f $COMPOSE_FILE -p $NAMESPACE up -d

echo "Running Go server..."
go run main.go

wait