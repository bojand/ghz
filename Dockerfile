# syntax=docker.io/docker/dockerfile:1.3-labs@sha256:250ce669e1aeeb5ffb892b18039c3f0801466536cb4210c8eb2638e628859bfd

FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.17 AS alpine
FROM --platform=$BUILDPLATFORM gcr.io/distroless/base:nonroot@sha256:e406b1da09bc455495417a809efe48a03c48011a89f6eb57b0ab882508021c0d AS distroless

FROM alpine AS osmap-linux
RUN echo linux   >/os
FROM alpine AS osmap-macos
RUN echo darwin  >/os
FROM alpine AS osmap-windows
RUN echo windows >/os
FROM osmap-$TARGETOS AS osmap

FROM alpine AS fetcher
WORKDIR /app
ARG VERSION
RUN \
    --mount=from=osmap,source=/os,target=/os <<EOF
set -eux
apk add --no-cache curl
export url=https://github.com/bojand/ghz/releases
export arch=x86_64
if [ "${VERSION:-}" = '' ]; then
    export VERSION=$( ( curl -#fSLo /dev/null -w '%{url_effective}' $url/latest && echo ) | while read -r x; do basename $x; done)
fi
curl -#fSLo exe.tar.gz $url/download/$VERSION/ghz-$(cat /os)-$arch.tar.gz
curl -#fSLo sha2 $url/download/$VERSION/ghz-$(cat /os)-$arch.tar.gz.sha256
sha256sum exe.tar.gz | grep -F $(cat sha2)
tar xvf exe.tar.gz
rm ghz-web* && mkdir exe && mv ghz* exe/
EOF

FROM scratch AS ghz-binary
COPY --from=fetcher /app/exe/* /

FROM distroless AS ghz
COPY --from=ghz-binary --chown=nonroot /ghz /
RUN ["/ghz", "--version"]
ENTRYPOINT ["/ghz"]
