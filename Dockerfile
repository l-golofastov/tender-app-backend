FROM golang:1.23.1 AS build

COPY ./ ./

COPY go.mod ./
COPY go.sum ./
RUN go mod download

RUN go build -o bin/app ./src/cmd/tender-app
CMD ["./bin/app"]

EXPOSE 8080
