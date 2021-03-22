FROM golang:1.16

# Set the Current Working Directory inside the container
WORKDIR /store

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependancies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

RUN go build -o ./app/store ./app/server.go

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the executable
CMD ["./apps/store run"]