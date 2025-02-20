FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./

RUN apk --no-cache add build-base

RUN go build -o agent .

FROM alpine:3.21.3

WORKDIR /app

COPY --from=builder /app/agent /app/

RUN apk --no-cache add curl

CMD ["./agent"]