#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../compose"
docker compose up -d
docker compose ps
echo "Mongo: mongodb://127.0.0.1:27017  db=talerole"
echo "Export MONGO_URI=mongodb://127.0.0.1:27017 MONGO_DB=talerole before starting the API."
