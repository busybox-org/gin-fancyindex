#docker build --network host --rm -t xmapst/gin-fancyindex:latest -f Dockerfile .
FROM golang:latest as builder

WORKDIR /go/src/gin-fancyindex
COPY . /go/src/gin-fancyindex

RUN go mod tidy && CGO_ENABLED=0 go build -ldflags "-w -s" -o fancyindex cmd/fancyindex.go


FROM alpine:latest
COPY --from=builder --chmod=0777 /go/src/gin-fancyindex/fancyindex /usr/local/bin/fancyindex
RUN apk add --no-cache ca-certificates mailcap

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/fancyindex"]
