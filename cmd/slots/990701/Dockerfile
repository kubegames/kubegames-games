FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 990701 /990701
ADD conf/ /conf
RUN chmod -R 777 /990701

WORKDIR /

ENTRYPOINT ./990701