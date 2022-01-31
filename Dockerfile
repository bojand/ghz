# syntax=docker.io/docker/dockerfile:1@sha256:42399d4635eddd7a9b8a24be879d2f9a930d0ed040a61324cfdf59ef1357b3b2

FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.15 AS alpine
FROM --platform=$BUILDPLATFORM gcr.io/distroless/base:nonroot@sha256:02f667185ccf78dbaaf79376b6904aea6d832638e1314387c2c2932f217ac5cb AS distroless

FROM alpine AS osmap-linux
RUN echo linux   >/os
FROM alpine AS osmap-macos
RUN echo darwin  >/os
FROM alpine AS osmap-windows
RUN echo windows >/os
FROM osmap-$TARGETOS AS osmap

FROM alpine AS fetcher
WORKDIR /app
RUN \
    --mount=from=osmap,source=/os,target=/os \
    set -ux \
 && apk add --no-cache curl \
 && export url=https://github.com/bojand/ghz/releases \
 && export arch=x86_64 \
 && export VERSION=$( ( curl -#fSLo /dev/null -w '%{url_effective}' $url/latest && echo ) | while read -r x; do basename $x; done) \
 && curl -#fSLo exe.tar.gz $url/download/$VERSION/ghz-$(cat /os)-$arch.tar.gz \
 && curl -#fSLo sha2 $url/download/$VERSION/ghz-$(cat /os)-$arch.tar.gz.sha256 \
 && sha256sum exe.tar.gz | grep -F $(cat sha2) \
 && tar xvf exe.tar.gz \
 && rm ghz-web* && mkdir exe && mv ghz* exe/

FROM scratch AS ghz-binary
COPY --from=fetcher /app/exe/* /

FROM distroless AS ghz
COPY --from=ghz-binary --chown=nonroot /ghz /
RUN ["/ghz", "--version"]
ENTRYPOINT ["/ghz"]
