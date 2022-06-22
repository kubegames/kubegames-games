FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960212 /960212
ADD conf/ /conf
RUN chmod -R 777 /960212

WORKDIR /

ENTRYPOINT ./960212