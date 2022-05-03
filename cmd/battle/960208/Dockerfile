FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960208 /960208
ADD conf/ /conf
RUN chmod -R 777 /960208

WORKDIR /

ENTRYPOINT ./960208