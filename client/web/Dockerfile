FROM nginx:1.15.0

ADD Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

ADD ./ /usr/share/nginx/html/

EXPOSE 80