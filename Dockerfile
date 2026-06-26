# ==========================================
# STAGE 1: Build the Go Application
# ==========================================
FROM golang:1.26-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
# (This ensures we don't redownload dependencies every time we change a .go file)
COPY go.mod go.sum ./
RUN go mod download

# Copy all the Go source code (your flat structure)
COPY . .

# Build the Go app as a static binary
# CGO_ENABLED=0 ensures it doesn't rely on the host OS's C libraries
# -o myapp outputs the compiled binary with the name "myapp"
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp .

# ==========================================
# STAGE 2: Create the minimal production image
# ==========================================
FROM alpine:latest

# Install CA certificates (required if your Go app makes HTTPS requests)
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/myapp .

# Expose port 8080 (Cloud Run's default port)
EXPOSE 8080

# Command to run the executable
CMD ["./myapp"]
