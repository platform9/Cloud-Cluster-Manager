FROM golang:1.17 as builder
WORKDIR /manager
COPY . .
RUN go build -o CloudManager
ENTRYPOINT ["./CloudManager"]