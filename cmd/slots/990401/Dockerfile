FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 990401 /990401
ADD conf/ /conf
RUN chmod -R 777 /990401

WORKDIR /

ENTRYPOINT ./990401