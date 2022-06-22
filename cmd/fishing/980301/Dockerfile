FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 980301 /980301
ADD conf/ /conf
RUN chmod -R 777 /980301

WORKDIR /

ENTRYPOINT ./980301