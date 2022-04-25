FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960101 /960101
ADD conf/ /conf
RUN chmod -R 777 /960101

WORKDIR /

ENTRYPOINT ./960101