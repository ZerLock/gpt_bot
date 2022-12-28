###
# Stage 1 - build
###
FROM golang:1.18-alpine as builder

WORKDIR /app

# Download dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy code
COPY . .

# Build the app
RUN go build -o /main

###
# Stage 2 - run
###
FROM alpine:3.17 as runner

# Copy executable
COPY --from=builder /main /main

# Set environment variables
ENV GPT_TOKEN="gpt-token"
ENV DISCORD_CLIENT_TOKEN="discord-token"

# Start server
ENTRYPOINT ["/main"]
