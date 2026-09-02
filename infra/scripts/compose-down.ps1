#Requires -Version 5
Set-Location (Join-Path $PSScriptRoot "..\compose")
docker compose down
