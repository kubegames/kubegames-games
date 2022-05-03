FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 960211 /960211
ADD conf/ /conf
RUN chmod -R 777 /960211

WORKDIR /

ENTRYPOINT ./960211