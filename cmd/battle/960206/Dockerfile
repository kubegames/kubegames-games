FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960206 /960206
ADD conf/ /conf
RUN chmod -R 777 /960206

WORKDIR /

ENTRYPOINT ./960206