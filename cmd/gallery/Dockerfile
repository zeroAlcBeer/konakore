FROM golang:alpine as builder

RUN apk add --no-cache make git

WORKDIR /konakore-src
COPY . /konakore-src
RUN export GOPROXY=https://goproxy.io,direct && \
    go mod download && \
    make docker && \
    mv ./bin/konakore /konakore

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/CheerChen/konakore"

COPY --from=builder /konakore /

ENTRYPOINT ["/konakore"]