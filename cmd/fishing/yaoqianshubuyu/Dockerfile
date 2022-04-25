FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD yaoqianshubuyu /yaoqianshubuyu
ADD config/ /config
RUN chmod -R 777 /yaoqianshubuyu

WORKDIR /

ENTRYPOINT ./yaoqianshubuyu