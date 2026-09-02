module github.com/leventkok/tale-role/apps/api

go 1.23.0

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/cors v1.2.1
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/uuid v1.6.0
	github.com/leventkok/tale-role/services/llm-gateway v0.0.0
	golang.org/x/crypto v0.36.0
)

replace github.com/leventkok/tale-role/services/llm-gateway => ../../services/llm-gateway
