FROM golang:1.23.10 as base

FROM base as dev

RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

WORKDIR /usr/src/app
COPY . .

# Install dependencies
RUN go mod download

# Expose worker port (if needed for health checks)
EXPOSE 8081

# Run the worker with air for hot reloading using worker config
CMD ["air", "-c", ".air.worker.toml"] 