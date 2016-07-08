FROM alpine

RUN apk update && apk add go git && mkdir /gopath
ENV GOPATH=/gopath
RUN go get github.com/aws/aws-sdk-go/aws && \
    go get github.com/prometheus/client_golang/prometheus

COPY . /ase
WORKDIR /ase
RUN go build
CMD ["/ase/ase"]
