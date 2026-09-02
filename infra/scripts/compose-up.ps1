#Requires -Version 5
Set-Location (Join-Path $PSScriptRoot "..\compose")
docker compose up -d
docker compose ps
Write-Host "Mongo: mongodb://127.0.0.1:27017  db=talerole"
Write-Host "Mailhog SMTP: 127.0.0.1:1025  UI: http://127.0.0.1:8025"
Write-Host "Then in apps/api: `$env:MONGO_URI='mongodb://127.0.0.1:27017'; `$env:MONGO_DB='talerole'; `$env:SMTP_HOST='127.0.0.1'; `$env:SMTP_PORT='1025'"
