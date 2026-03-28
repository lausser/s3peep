# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /s3peep ./cmd/s3peep

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /home/s3peep

COPY --from=builder /s3peep .
COPY web/ ./web/

EXPOSE 8080

ENTRYPOINT ["/home/s3peep"]
CMD ["serve", "--port", "8080"]
