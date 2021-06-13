FROM golang:1.16.5-buster
ENV GO111MODULE=on
RUN mkdr /app
ADD . /app
WORKDIR /app
RUN gobuild -0 main .
CMD ["/app/main"]