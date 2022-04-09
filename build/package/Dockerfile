FROM alpine:3.14
MAINTAINER mysql2elasticsearch <https://github.com/zhangjunjie6b/mysql2elasticsearch>
COPY ./bin /root/bin
EXPOSE 9102
WORKDIR /root/bin
RUN chmod 777 ./main_linux64
CMD ./main_linux64