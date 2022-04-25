FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960203 /960203
ADD conf/ /conf
RUN chmod -R 777 /960203

WORKDIR /

ENTRYPOINT ./960203