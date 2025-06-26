FROM golang:1.23.10 as base

FROM base as dev

RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

WORKDIR /usr/src/app
COPY . .

# Install dependencies
RUN go mod download

# Build the API binary
RUN go build -o bin/api bootstrap/api/main.go

EXPOSE 8080
CMD ["air"] 