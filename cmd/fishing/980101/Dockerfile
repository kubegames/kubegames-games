FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 980101 /980101
ADD conf/ /conf
RUN chmod -R 777 /980101

WORKDIR /

ENTRYPOINT ./980101