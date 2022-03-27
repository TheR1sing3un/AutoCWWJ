FROM golang:1.17.6-alpine3.15 as build
LABEL maintainer="ther1sing3un"
ENV GO111MODULE=on
WORKDIR /root/project
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix cgo -o server_bin
FROM alpine:3.15 as prod
EXPOSE 6666
ENV TZ Asia/Shanghai
RUN apk add tzdata && cp /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone
WORKDIR /root
COPY --from=build /root/project/server_bin /root/server_bin
COPY --from=build /root/project/config.yaml /root/config.yaml
ENTRYPOINT /root/server_bin