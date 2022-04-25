FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960204 /960204
ADD conf/ /conf
RUN chmod -R 777 /960204

WORKDIR /

ENTRYPOINT ./960204