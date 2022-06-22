FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 990201 /990201
ADD conf/ /conf
RUN chmod -R 777 /990201

WORKDIR /

ENTRYPOINT ./990201