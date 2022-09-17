FROM fangdajiang/centos:centos7.9.2022.07

MAINTAINER "fangdajiang@gmail.com"
LABEL description="YouTube 视频转音频"

ENV BOT_TOKEN ${BOT_TOKEN}
ENV BOT_CHAT_ID ${BOT_CHAT_ID}
ENV CHAT_ID ${CHAT_ID}
ENV YOUTUBE_KEY ${YOUTUBE_KEY}
ENV ALICLOUD_ACCESS_KEY ${ALICLOUD_ACCESS_KEY}
ENV ALICLOUD_SECRET_KEY ${ALICLOUD_SECRET_KEY}
ENV ALICLOUD_REGION ${ALICLOUD_REGION}
ENV DOCKER_ID ${DOCKER_ID}

ENV LANG     en_US.UTF-8
ENV LANGUAGE en_US.UTF-8
ENV LC_ALL   en_US.UTF-8
RUN /bin/cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'Asia/Shanghai' >/etc/timezone

ADD bin/dependency/youtube-dl /usr/local/sbin/
ADD bin/ya /app/ya

ENTRYPOINT ["/app/ya", "run", "-m", "all"]