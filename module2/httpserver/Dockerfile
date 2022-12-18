#辅助构建
FROM golang:1.17 AS builder
ENV GO111MODULE=off \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build
COPY . .
RUN go build -o httpserver .


FROM busybox
COPY --from=builder /build/httpserver /
EXPOSE 8080
ENTRYPOINT ["/httpserver"]
