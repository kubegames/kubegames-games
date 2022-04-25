FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 940101 /940101
ADD conf/ /conf
RUN chmod -R 777 /940101

WORKDIR /

ENTRYPOINT ./940101