## 结合 Telegram 的 机器人(Bot) 和 频道(Channel)，将视频网站（目前仅 YouTube）上的订阅内容自动以音频的形式发送到 Telegram 的指定频道中，从而提供 Telegram 上的音频服务。

#### 简体中文 [English](/docs/en_US/README.md)

## 演示
> * 注册 Telegram（俗称电报，中国大陆须科学上网，或使用其内置的代理）
> * 在 Telegram 中订阅频道 @YouTubeCnPoliticsAudio
> * 从 Telegram 不定时获得 YouTuber(s) 发布的音频内容（因 Telegram 的限制，不拉取大小超过 50M 的音轨）
> * 在 Telegram 的机器人 @you_audio_bot 中查看告警通知

## 构建
> * 依赖
>   * Go 1.17+
>   * Python 2.7.5+
>   * Docker
>   * Packer
>   * (Terraform, [Linux 仓库设置](https://www.hashicorp.com/blog/announcing-the-hashicorp-linux-repository))
>   * OSS ([youtube-audio/fetch_base.json](https://youtube-audio.oss-cn-hongkong.aliyuncs.com/fetch_base.json) 和 [youtube-audio/fetch_history.json](https://youtube-audio.oss-cn-hongkong.aliyuncs.com/fetch_history.json))
>   * ```shell
>     sudo curl -L https://yt-dl.org/downloads/latest/youtube-dl -o /usr/local/bin/youtube-dl
>     sudo chmod a+rx /usr/local/bin/youtube-dl
>     ```
>   * 建立 Telegram 机器人 和 频道，并将该机器人加入到频道中并设为管理员
> * 从源码构建
> ```shell
> git clone https://github.com/fangdajiang/youtube-audio.git
> cd youtube-audio
> go build -o bin/ya main.go
> ```
> * 通过 Docker 构建
> ```shell
> docker build -t youtube-audio:latest -f ./Dockerfile .
> ```
> * 通过 Packer 构建并推到 Docker Hub 中
> ```shell
> # 设置环境变量 DOCKER_ID
> packer build deploy/packer/local.json
> ```
> * 通过 Terraform 来 Provision
> ```shell
> # 设置环境变量 ALICLOUD_ACCESS_KEY, ALICLOUD_SECRET_KEY, ALICLOUD_REGION
> # 
> # 先在阿里云构建基础镜像，获取所生成的镜像 ID
> packer build deploy/packer/alicloud.json
> # 将该 ID 更新到 main.tf 的 image_id
> ```

## 运行
> * Docker
> ```shell
> docker run -d -e BOT_TOKEN= -e BOT_CHAT_ID= -e CHAT_ID= -e YOUTUBE_KEY= -e ALICLOUD_ACCESS_KEY= -e ALICLOUD_SECRET_KEY= youtube-audio:latest
> ```
> * Terraform
> ```shell
> cd deploy/terraform/alicloud
> terraform init/plan/apply
> ```
> * Dev
>   * 拷贝 bin/dependency/youtube-dl 到本机 $PATH
>   * 设置本机环境变量 BOT_TOKEN, BOT_CHAT_ID, CHAT_ID, YOUTUBE_KEY
> ```shell
> # 拉取近期音频
> go run main.go run -m latest
> # 拉取单条音频
> go run main.go run -m single https://www.youtube.com/watch?v=xxx
> ```

## 功能
- [x] CLI 支持一键拉取自定义 YouTuber Playlist 近期发布的 2 条视频的音轨到 Telegram 的指定频道
- [x] 支持 Packer 在阿里云平台上构建镜像
- [x] 支持 Terraform 在阿里云平台上构建虚拟机
- [x] CLI 支持拉取单条视频的音轨到 Telegram 指定频道
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
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/ya main.go
```
> *

## 受以下项目启发并表示感谢
* [youtube-dl](https://github.com/ytdl-org/youtube-dl)
* [youtube](https://github.com/kkdai/youtube)
* [YouTube中文時政精選](https://t.me/YouTubePoliTalk)
