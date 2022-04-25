FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960202 /960202
ADD conf/ /conf
RUN chmod -R 777 /960202

WORKDIR /

ENTRYPOINT ./960202