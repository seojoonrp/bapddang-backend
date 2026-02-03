#!/bin/bash

echo "Killing existing process on port 8080..."
lsof -ti:8080 | xargs kill -9 2>/dev/null

echo "Starting ngrok..."
ngrok http 8080 --log=stdout > /dev/null &
NGROK_PID=$!

echo "Starting Go server..."
go run main.go

trap "kill $NGROK_PID" EXIT