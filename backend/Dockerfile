# Build stage
FROM golang:1.24-alpine as builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o tvseries .

# Runtime stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/tvseries .
COPY --from=builder /app/docs ./docs

EXPOSE 8080
CMD ["./tvseries"]
