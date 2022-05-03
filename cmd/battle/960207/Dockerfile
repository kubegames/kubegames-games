FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960207 /960207
ADD conf/ /conf
RUN chmod -R 777 /960207

WORKDIR /

ENTRYPOINT ./960207