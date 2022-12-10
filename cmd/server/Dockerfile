FROM golang:alpine as builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add --no-cache make git

WORKDIR /konakore-src
COPY . /konakore-src

RUN make server && \
    mv ./bin/server /server

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/CheerChen/konakore"

COPY --from=builder /server /

ENTRYPOINT ["/server"]