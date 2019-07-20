FROM golang AS build-env

ENV GO111MODULE=on

ADD . /go/src

WORKDIR /go/src

RUN go mod vendor

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o /go/bin/throttler .

FROM scratch

COPY --from=build-env /go/bin/throttler /go/bin/throttler

CMD ["/go/bin/throttler"]
