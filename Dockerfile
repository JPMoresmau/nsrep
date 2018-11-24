FROM golang:1.8-alpine

WORKDIR /go/src/github.com/JPMoresmau/metarep

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["metarep"]