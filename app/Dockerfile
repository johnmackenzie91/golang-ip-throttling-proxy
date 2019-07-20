FROM golang AS build-env

#ENV GO111MODULE=on

ADD . /go/src

WORKDIR /go/src

#RUN go mod vendor

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/app .

FROM scratch

COPY --from=build-env /go/bin /go/bin

CMD ["/go/bin/app"]
