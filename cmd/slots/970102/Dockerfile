FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 970102 /970102
ADD conf/ /conf
RUN chmod -R 777 /970102

WORKDIR /

ENTRYPOINT ./970102