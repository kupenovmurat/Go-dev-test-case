#!/bin/bash

mkdir -p data/rest
mkdir -p data/storage-{1..6}
mkdir -p test-data
mkdir -p logs

go run ./cmd/rest-server/main.go --port=8080 > logs/rest.log 2>&1 &
echo "Started REST server on port 8080"

sleep 2

for i in {1..6}; do
  go run ./cmd/storage-server/main.go --port=808$i --id=storage-$i --rest=http://localhost:8080 --data=./data/storage-$i > logs/storage-$i.log 2>&1 &
  echo "Started storage server $i on port 808$i"
done

echo "All servers started. Use 'pkill -f \"go run ./cmd\"' to stop all servers." 