FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960302 /960302
ADD conf/ /conf
RUN chmod -R 777 /960302

WORKDIR /

ENTRYPOINT ./960302