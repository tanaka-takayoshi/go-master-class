FROM golang:1.14

RUN apt-get update && apt-get -y install mariadb-client
ENV GO111MODULE=on

WORKDIR /go/src/webapp
CMD ["go", "run", "main.go"]