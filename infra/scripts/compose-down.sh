#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../compose"
docker compose down
