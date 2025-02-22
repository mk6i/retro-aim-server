# Specifies a parent image
FROM golang:alpine
 
# Creates an app directory to hold your app’s source code
WORKDIR /app
 
# Copies everything from your root directory into /app
COPY . .
COPY config/settings.env config/settings.env
 
# Installs Go dependencies
RUN go mod download
 
# Builds your app with optional configuration
RUN go build -o retro_aim_server ./cmd/server
 
# Tells Docker which network port your container listens on
EXPOSE 8080 5194 5190 5195 5191 5193 5192 5196 5197
 
# Specifies the executable command that runs when the container starts
CMD ["/app/retro_aim_server", "-config", "config/settings.env"]
