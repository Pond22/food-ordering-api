# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

# Install tzdata in final stage
RUN apk add --no-cache tzdata

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .
# Copy font file
COPY --from=builder /app/THSarabunNew.ttf .
COPY --from=builder /app/logo.jpg .

COPY dict/lexitron.txt /app/dict/lexitron.txt

# Command to run the application
CMD ["./main"]