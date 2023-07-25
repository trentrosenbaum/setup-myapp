FROM golang:1.20.6

# Install Git and other dependencies
RUN apt-get update && apt-get install -y \
    git

# Set working directory
WORKDIR /github/workspace

# Copy the Go code to the container
COPY . .

# Build the Go program
RUN go build -o setup-app-action

# Set the entrypoint for the action
ENTRYPOINT ["/github/workspace/setup-app-action"]
