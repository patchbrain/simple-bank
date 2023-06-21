FROM golang:1.18.10-alpine3.17 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
# RUN apk add curl
# RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz

FROM alpine:3.17
WORKDIR /app
# 若不加该语句，当执行wait-for.sh脚本时，无法找到bash程序
RUN apk add --no-cache bash
COPY --from=builder /app/main .
# COPY --from=builder /app/migrate .
COPY app.env .
# 复制迁移文件
COPY db/migration ./db/migration
COPY start.sh .
COPY wait-for.sh .
RUN chmod +x /app/start.sh
RUN chmod +x /app/wait-for.sh

# EXPOSE仅仅作为镜像创建者之间的交流
EXPOSE 8080
CMD ["/app/main"]
ENTRYPOINT ["/app/start.sh"]