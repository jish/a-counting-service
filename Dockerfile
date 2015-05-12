FROM golang

ADD . /go/src/app
RUN cd /go/src/app && go get

ENTRYPOINT /go/bin/app

EXPOSE 3000
