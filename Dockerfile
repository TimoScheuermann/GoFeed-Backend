# Build stage
FROM golang:alpine AS builder

RUN apk --no-cache add gcc g++ make git

WORKDIR /app
COPY . .
RUN GOOS=linux go build -ldflags="-s -w" -o main .

# Run stage
FROM alpine:3.13
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 80
ENTRYPOINT /app/main --port 80
