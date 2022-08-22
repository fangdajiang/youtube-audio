## 结合 Telegram 的 机器人(Bot) 和 频道(Channel)，将视频网站（目前仅 YouTube）上的订阅内容自动以音频的形式发送到 Telegram 的指定频道中，从而提供 Telegram 上的音频服务。

#### 简体中文 [English](/docs/en_US/README.md)

## 演示
> * 注册 Telegram（俗称电报，中国大陆须科学上网，或使用其内置的代理）
> * 在 Telegram 中订阅频道 @YouTubeCnPoliticsAudio
> * 不定时获得某 YouTuber 发布的音频内容
> * 在 Telegram 的机器人 @you_audio_bot 中查看告警通知

## 构建
> * 依赖
    -> Go 1.17+
    -> Python 2.7.5+
    -> Packer
    -> (Docker, Terraform, [Linux 仓库设置](https://www.hashicorp.com/blog/announcing-the-hashicorp-linux-repository))
```shell
sudo curl -L https://yt-dl.org/downloads/latest/youtube-dl -o /usr/local/bin/youtube-dl
sudo chmod a+rx /usr/local/bin/youtube-dl
```
> * 从源码安装
```shell
# 拷贝 bin/dependency/youtube-dl 到 $PATH
# 设置环境变量 BOT_TOKEN, BOT_CHAT_ID, CHAT_ID, YOUTUBE_KEY
git clone https://github.com/fangdajiang/youtube-audio.git
cd youtube-audio
go run ./cmd/main.go
```
> * 通过 Packer(Docker) 安装并推到 Docker Hub 中
```shell
# 设置环境变量 DOCKER_ID
packer build deploy/packer/local.json
```
> * 通过 Terraform 安装
```shell
# 还须设置环境变量 ALICLOUD_ACCESS_KEY, ALICLOUD_SECRET_KEY, ALICLOUD_REGION
packer build deploy/packer/alicloud.json
```

## 例子
Docker:
```shell
docker run -d -e BOT_TOKEN= -e BOT_CHAT_ID= -e CHAT_ID= -e YOUTUBE_KEY= -e ALICLOUD_ACCESS_KEY= -e ALICLOUD_SECRET_KEY= youtube-audio:latest
```
Terraform:
```shell
# 在云平台上获取所生成的镜像 ID
cd deploy/terraform
terraform init/plan/apply
```

## 功能
- [x] 一键拉取 YouTuber 最近发布的 1 条音频内容到 Telegram 的指定频道
- [x] 支持 Packer 在阿里云平台上构建镜像
- [x] 支持 Terraform 在阿里云平台上构建虚拟机
- [ ] 使用 Bot 进行订阅
- [ ] 略延时获取发布内容的音频（由于 YouTube 会延时发布视频中的单独音轨，故暂时无法做到实时）
- [ ] 支持订阅多 YouTuber
- [ ] 支持取消订阅
- [ ] 支持发布给用户
- [ ] （支持订阅不同质量的音轨）
- [ ] 将本项目平台化，订阅来源和发布目的 与本项目解耦

## 注意事项
> * 编译时，假如本机是 ARM 架构的 CPU(如 Apple M1/M2)，须加参数
```shell
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/youtube-audio cmd/main.go
docker build -t youtube-audio:latest -f ./Dockerfile .
```
> *

## 受以下项目启发并表示感谢
* [youtube-dl](https://github.com/ytdl-org/youtube-dl)
* [youtube](https://github.com/kkdai/youtube)
