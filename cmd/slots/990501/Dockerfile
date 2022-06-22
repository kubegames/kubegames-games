FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 990501 /990501
ADD conf/ /conf
RUN chmod -R 777 /990501

WORKDIR /

ENTRYPOINT ./990501