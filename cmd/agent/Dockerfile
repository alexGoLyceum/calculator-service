FROM golang:1.23 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/agent cmd/agent/main.go

FROM debian:bookworm-slim

COPY --from=builder /bin/agent /bin/agent

WORKDIR /bin

ENTRYPOINT ["/bin/agent"]