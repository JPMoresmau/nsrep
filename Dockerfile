FROM metarep:build


WORKDIR /go/src/github.com/JPMoresmau/metarep

COPY . .

RUN go install -v ./...


CMD ["metarep"]