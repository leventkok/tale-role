#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../compose"
docker compose up -d
docker compose ps
echo "Mongo: mongodb://127.0.0.1:27017  db=talerole"
echo "Mailhog SMTP: 127.0.0.1:1025  UI: http://127.0.0.1:8025"
echo "Export MONGO_URI=mongodb://127.0.0.1:27017 MONGO_DB=talerole SMTP_HOST=127.0.0.1 SMTP_PORT=1025 before starting the API."
