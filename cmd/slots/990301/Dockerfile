FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 990301 /990301
ADD conf/ /conf
RUN chmod -R 777 /990301

WORKDIR /

ENTRYPOINT ./990301