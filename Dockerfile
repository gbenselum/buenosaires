# ---- Builder Stage ----
# Use an official Go image as a parent image
FROM golang:1.24-alpine AS builder

# Set the necessary environment variables for a static build
ENV CGO_ENABLED=0
ENV GOOS=linux

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files to leverage Docker cache
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o /buenosaires main.go


# ---- Final Stage ----
# Use a lightweight alpine image for the final container
FROM alpine:latest

# Install runtime dependencies required by the application and plugins
# git: for repository operations
# bash: for executing shell scripts
# shellcheck: for linting shell scripts
# sudo: to allow scripts to run with elevated privileges if configured
RUN apk add --no-cache git bash shellcheck sudo

# Copy the pre-built binary from the builder stage
COPY --from=builder /buenosaires /usr/local/bin/buenosaires

# Create a non-root user and group for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Allow the non-root user to run sudo commands without a password prompt.
# This is necessary for the shell plugin's `allow_sudo` feature to work
# in a non-interactive container environment.
RUN echo "appuser ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

# Switch to the non-root user
USER appuser

# Set the working directory in the container
WORKDIR /app

# Set the entrypoint for the container. When the container runs, it will execute the 'buenosaires' binary.
# Users can then pass commands like 'install' or 'run' to the container.
ENTRYPOINT ["buenosaires"]