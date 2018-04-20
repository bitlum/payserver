FROM ubuntu:16.04
MAINTAINER Samokhvalov Andrey <andrey@bitlum.io>

# Install root certificates so that we could fetch https links.
RUN apt-get update
RUN apt-get install -y ca-certificates
RUN update-ca-certificates

WORKDIR /

# This implies that service has to be built locally first, and putted in the
# docker directory, fore running docker build.
COPY connector .

# That is needed to have communication with lightning network daemon.
COPY lnd-tls.cert ./lnd-tls.cert

# This implies that service config is defines and located in the
# directory with docker file.
COPY connector.conf .

ENTRYPOINT [ "./connector", "--config", "connector.conf"]