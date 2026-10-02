FROM alpine:3.24

RUN apk add --no-cache \
    libreoffice \
    openjdk25-jre-headless \
    pandoc-cli \
    ffmpeg \
    fontconfig \
    ttf-dejavu

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/ccmb /usr/bin/ccmb

ENV HOME=/tmp

USER nobody
ENTRYPOINT ["/usr/bin/ccmb"]
