FROM golang:1.13.5-buster

MAINTAINER Roma Erema

COPY . /wapi_build

ENV GOPATH /wapi_build

RUN echo "building wapi..."
RUN cd /wapi_build \
    && go build -o /wapi/build main.go \
    && rm -rf /wapi_build \
RUN echo "ok"

WORKDIR /wapi

CMD ./build
