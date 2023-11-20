# Stage 1: Compile the binary in a containerized Golang environment
#
FROM golang:1.19.2-alpine as build
# Set the working directory to the same place we copied the code
WORKDIR /src
# Copy the source files from the host
COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

# Build the binary!
#RUN CGO_ENABLED=0 GOOS=linux go build -a -o melon cmd/melon/main.go
RUN go build -o melon cmd/melon/main.go


# Stage 2: Build the Key-Value Store image proper
#
# Use a "scratch" image, which contains no distribution files
FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN apk add bash
# Copy the binary from the build container
COPY --from=build /src .
## If you're using TLS, copy the .pem files too
#COPY --from=build /src/*.pem .
# Tell Docker we'll be using port 8080
EXPOSE 8001
# Tell Docker to execute this command on a "docker run"
CMD ["./melon"]
