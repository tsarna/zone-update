FROM golang:alpine AS builder
MAINTAINER Ty Sarna <ty@sarna.org>

RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .

RUN go get -d -v

RUN go build -o /zoneupdated

FROM scratch

COPY --from=builder /zoneupdated /zoneupdated

ENTRYPOINT ["/zoneupdated"]
