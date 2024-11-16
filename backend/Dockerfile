# Use official Go image
FROM golang:1.22

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./
COPY go.sum ./

# Download Go modules
RUN go mod download

# Copy the rest of the backend code
COPY . .

# Expose the port your app runs on
EXPOSE 4567

# Command to build and run the application
RUN go build -o backend .
CMD ["./backend"]
