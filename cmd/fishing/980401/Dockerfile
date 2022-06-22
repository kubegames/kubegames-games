FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 980401 /980401
ADD conf/ /conf
RUN chmod -R 777 /980401

WORKDIR /

ENTRYPOINT ./980401