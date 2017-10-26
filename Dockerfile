FROM golang:1.9 AS builder

WORKDIR /go/src/github.com/melnikk/kafka-collector
COPY . /go/src/github.com/melnikk/kafka-collector
RUN go get github.com/kardianos/govendor
RUN govendor sync
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/melnikk/kafka-collector/kafka-collector .

ENTRYPOINT ["/root/kafka-collector"]
