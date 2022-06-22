FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 980201 /980201
ADD conf/ /conf
RUN chmod -R 777 /980201

WORKDIR /

ENTRYPOINT ./980201