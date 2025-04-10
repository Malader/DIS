FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY worker/ ./worker/

WORKDIR /app/worker/cmd
RUN go build -o worker .

EXPOSE 8081

CMD ["/app/worker/cmd/worker"]
