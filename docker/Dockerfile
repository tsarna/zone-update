FROM golang:alpine AS builder
MAINTAINER Ty Sarna <ty@sarna.org>

ARG goos=linux
ARG goarch=amd64

RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .

RUN go get -d -v

RUN GOOS=${goos} GOARCH=${goarch} go build -ldflags="-w -s" -o /zoneupdated     

FROM scratch

COPY --from=builder /zoneupdated /zoneupdated

ENTRYPOINT ["/zoneupdated"]
