FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960201 /960201
ADD conf/ /conf
RUN chmod -R 777 /960201

WORKDIR /

ENTRYPOINT ./960201