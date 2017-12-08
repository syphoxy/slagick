FROM golang:1.8-alpine
WORKDIR /go/src/app
RUN apk --no-cache add git
RUN go get \
    github.com/lib/pq \
    github.com/nlopes/slack
COPY src /go/src/app
RUN go build slagick.go
CMD ["./slagick"]
