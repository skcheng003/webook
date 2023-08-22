# 基础镜像
FROM ubuntu:20.04
COPY webook /app/webook
# 自定义工作目录
WORKDIR /app
ENTRYPOINT ["/app/webook"]