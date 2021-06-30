# Build stage
FROM golang:alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o main .

# Run stage
FROM alpine:3.13
WORKDIR /app

COPY --from=builder /app/main .
COPY auth.html .

EXPOSE 3000
CMD ["/app/main"]
