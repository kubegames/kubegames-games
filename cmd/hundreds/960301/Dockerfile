FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960301 /960301
ADD conf/ /conf
RUN chmod -R 777 /960301

WORKDIR /

ENTRYPOINT ./960301