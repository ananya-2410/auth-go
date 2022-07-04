FROM golang:1.18.3
RUN mkdir -p / server
WORKDIR /server
ADD . /server
RUN go build ./server.go
CMD ["./server"]