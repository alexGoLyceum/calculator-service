FROM golang:1.23 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/orchestrator cmd/orchestrator/main.go

FROM debian:bookworm-slim

RUN mkdir -p /orchestrator/migrations

COPY --from=builder /bin/orchestrator /bin/orchestrator
COPY --from=builder /src/orchestrator/migrations /orchestrator/migrations

WORKDIR /bin

ENTRYPOINT ["/bin/orchestrator"]