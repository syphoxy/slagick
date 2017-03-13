FROM golang:1.8-alpine

RUN apk --no-cache add git \
  && mkdir -p /go/src/app
COPY src /go/src/app
WORKDIR /go/src/app
RUN go get github.com/lib/pq github.com/nlopes/slack \
  && go build slagick.go

CMD ["./slagick"]
