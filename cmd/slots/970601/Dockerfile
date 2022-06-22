FROM busybox:latest

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD 970601 /970601
ADD conf/ /conf
RUN chmod -R 777 /970601

WORKDIR /

ENTRYPOINT ./970601