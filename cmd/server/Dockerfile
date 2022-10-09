FROM golang:alpine as builder

RUN apk add --no-cache make git

WORKDIR /konakore-src
COPY . /konakore-src

ENV GO111MODULE=on \
    GOPROXY=https://proxy.golang.com.cn,direct

RUN make server && \
    mv ./bin/server /server

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/CheerChen/konakore"

COPY --from=builder /server /

ENTRYPOINT ["/server"]