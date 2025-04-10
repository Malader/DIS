FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY manager/ ./manager/

WORKDIR /app/manager/cmd
RUN go build -o manager .

EXPOSE 8080

CMD ["/app/manager/cmd/manager"]