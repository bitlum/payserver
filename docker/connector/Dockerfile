FROM alpine:3.7
MAINTAINER Samokhvalov Andrey <andrey@bitlum.io>

WORKDIR /

COPY connector .
COPY lnd-tls.cert .
COPY simnet-connector.conf .

ENTRYPOINT [ "./connector" ]