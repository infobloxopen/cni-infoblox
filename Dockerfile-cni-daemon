FROM golang:1.7-alpine as builder

RUN apk --update add --no-cache --virtual .build-deps \
    gcc libc-dev linux-headers

ENV SRC=/go/src/github.com/infobloxopen/cni-infoblox

COPY . ${SRC}
WORKDIR ${SRC}

RUN go build -o bin/infoblox-cni-daemon ./daemon


FROM alpine:3.5

ENV SRC=/go/src/github.com/infobloxopen/cni-infoblox
COPY --from=builder ${SRC}/bin/infoblox-cni-daemon /usr/local/bin/infoblox-cni-daemon

ENTRYPOINT ["/usr/local/bin/infoblox-cni-daemon"]

ARG GIT_SHA
ARG BUILD_DATE

LABEL GIT_SHA=$GIT_SHA \
      BUILD_DATE=$BUILD_DATE