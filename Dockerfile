FROM golang:1.18.10-alpine3.17 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/main .
#EXPOSE仅仅作为镜像创建者之间的交流
EXPOSE 8080
CMD ["/app/main"]