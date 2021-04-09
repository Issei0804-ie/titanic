FROM golang:1.15

WORKDIR /go/src/app
COPY . .
RUN chmod 0777 /go/src/app/cron.sh

# apt アップデートとcronのインストール
RUN apt clean -y
RUN apt update -y
RUN apt upgrade -y
RUN apt install -y sudo cron

# Dockerfileと同じ階層の"cron.d"フォルダ内にcronの処理スクリプトを格納しておく
ADD cron.d /etc/cron.d/
RUN chmod 0644 /etc/cron.d/*
RUN crontab /etc/cron.d/*

RUN go build -o main .
RUN touch /var/log/cron.log


CMD cron && tail -f /var/log/cron.log