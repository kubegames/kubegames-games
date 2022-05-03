FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960205 /960205
ADD conf/ /conf
RUN chmod -R 777 /960205

WORKDIR /

ENTRYPOINT ./960205