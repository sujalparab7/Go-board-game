# STAGE 1: Build the Binary
# Use the official Golang image to build the app
FROM golang:alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first (for better caching)
COPY go.mod ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
# -o main: names the output binary "main"
# CGO_ENABLED=0: ensures a static binary (important for Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# STAGE 2: Run the Binary
# Use a tiny image for the final container
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/main .

# Copying index.html to the final stage
COPY index.html .

# Expose the port your app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]