FROM alpine:3.24

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/ccmb /usr/bin/ccmb

USER nobody
ENTRYPOINT ["/usr/bin/ccmb"]
