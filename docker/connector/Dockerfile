FROM alpine:3.7
MAINTAINER Samokhvalov Andrey <andrey@bitlum.io>

# Install certificates to be able to make requests which requires TLS
# connection.
RUN apk --no-cache add openssl ca-certificates

WORKDIR /

# This implies that service has to be built locally first, and putted in the
# docker directory, fore running docker build.
COPY connector .

# That is needed to have communication with lightning network daemon.
COPY lnd-tls.cert .

# This implies that service config is defines and located in the
# directory with docker file.
COPY simnet-connector.conf .

ENTRYPOINT [ "./connector" ]