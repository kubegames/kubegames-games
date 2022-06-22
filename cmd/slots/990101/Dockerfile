FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 990101 /990101
ADD conf/ /conf
RUN chmod -R 777 /990101

WORKDIR /

ENTRYPOINT ./990101