# syntax=docker/dockerfile:1
FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY services/image-worker ./services/image-worker
COPY services/llm-gateway ./services/llm-gateway
COPY apps/api ./apps/api
WORKDIR /src/apps/api
RUN CGO_ENABLED=0 go build -tags netgo -ldflags "-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
ENV SERVER_HOST=0.0.0.0
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
