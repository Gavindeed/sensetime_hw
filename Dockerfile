FROM registry.sensetime.com/sre/golang:1.8.6-alpine3.6

ADD . /go/src/myftp

RUN go install myftp

ENTRYPOINT /go/bin/myftp

EXPOSE 2121-2200
