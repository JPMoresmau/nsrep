FROM nsrep:build


WORKDIR /go/src/github.com/JPMoresmau/nsrep

COPY . .

RUN go install -v ./...


CMD ["nsrep"]