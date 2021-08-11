FROM --platform=linux/arm64 golang:alpine as builder

RUN apk add --no-cache make git

WORKDIR /konakore-src
COPY . /konakore-src
RUN export GOPROXY=https://goproxy.io,direct && \
    make server && \
    mv ./bin/server /server

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/CheerChen/konakore"

COPY --from=builder /server /

ENTRYPOINT ["/server"]