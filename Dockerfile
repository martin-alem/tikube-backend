# Stage 1: Build the application
# Use the official Golang image for the building stage
FROM golang:1.21 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy Everything
COPY . .

# Download module dependencies
RUN go mod download

# Build the application
# -o logger sets the output name of the binary
RUN CGO_ENABLED=0 GOOS=linux go build -v -o logger ./cmd/main/main.go

# Expose port (adjust if different)
EXPOSE 8080

# Run the binary
CMD ["./logger"]
