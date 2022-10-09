FROM golang:alpine as builder

RUN apk add --no-cache make git

WORKDIR /konakore-src
COPY . /konakore-src

ENV GO111MODULE=on \
    GOPROXY=https://proxy.golang.com.cn,direct

RUN make syncer && \
    mv ./bin/syncer /syncer

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/CheerChen/konakore"

COPY --from=builder /syncer /

ENTRYPOINT ["/syncer"]