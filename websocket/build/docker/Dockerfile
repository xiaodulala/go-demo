FROM docker.hgpark.cn/yftest/duyong/ubuntu:latest
MAINTAINER duyong<tyduyong@126.com>

# 设置工作目录
ENV MYPATH=/duyong
ENV CONFIG_PATH=$MYPATH/configs

# 创建文件夹
RUN mkdir -p $CONFIG_PATH

# 拷贝可执行文件
COPY ./tmp/wssvc  $MYPATH

# 拷贝配置文件
ADD ./tmp/configs/ $CONFIG_PATH/

# 映射配置文件和日志
VOLUME ["$CONFIG_PATH"]

# 容器监听端口
EXPOSE 9099


WORKDIR $MYPATH

#启动
ENTRYPOINT ["./wssvc"]
CMD ["-c","./configs/apiserver.yaml"]


