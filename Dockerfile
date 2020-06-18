FROM golang:1.14 AS builder
WORKDIR /var/tmp/wapi
COPY . /var/tmp/wapi
RUN CGO_ENABLED=0 GOOS=linux go build -o /var/tmp/wapi/bin/wapi /var/tmp/wapi/main.go

FROM alpine:3.12
MAINTAINER Roma Erema
WORKDIR /root/
COPY --from=builder /var/tmp/wapi/bin/wapi .
COPY ./.env .
CMD ["./wapi"]
