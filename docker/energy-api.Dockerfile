FROM golang:1.15.6-alpine
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go build -v -o ./energy ./cmd/energy-api/main.go
CMD ["./energy"]