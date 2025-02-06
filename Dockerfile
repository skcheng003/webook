# 基础镜像
FROM ubuntu:20.04
COPY webook /bin/webook
# 自定义工作目录
WORKDIR /bin
ENTRYPOINT ["/bin/webook"]
