FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./

RUN go build -o datafrog-runner-agent ./cmd/agent

FROM alpine:3.21.3

ENV MONITOR_INTERVAL="30"

WORKDIR /app

COPY --from=builder /app/datafrog-runner-agent .

CMD ["./datafrog-runner-agent"]