# Repository Guidelines

## 项目结构与模块组织
- CLI 入口位于 `main.go`，不同子命令集中在 `cmd/`（例如 `cmd/run.go` 负责拉流任务，`cmd/ya.go` 定义全局 flag）；若新增命令，请保持与现有 `cobra` 风格一致。
- 核心业务归档在 `pkg/`：`handler/` 负责从 YouTube 拉取音轨并投递至 Telegram，`reporter/` 汇总处理结果，`util/` 提供数据库、OSS、日志等基础设施工具；新增模块时优先放在相应子包以维持职责清晰。
- 预置二进制存放在 `bin/`，其中 `bin/dependency/` 包含版本固定的 `yt-dlp`、`ffmpeg` 等依赖，更新时务必说明来源和校验方式；`bin/ya` 为默认构建产物。
- 部署脚本与模板位于 `deploy/`，分别涵盖 Docker 镜像、Packer 镜像构建与 Terraform 上云流程；中文/英文文档位于 `README.md` 与 `docs/en_US/README.md`，`logs/` 保存运行样例日志，便于排查。

## 构建、测试与本地开发
- `go build -o bin/ya main.go`：生成本地可执行文件；如在 Apple Silicon 上交叉编译请附加 `CGO_ENABLED=0 GOOS=linux GOARCH=amd64`，同时确认 `yt-dlp` 已安装在 `$PATH`。
- `go run main.go run -m latest`：拉取订阅播放列表最近的音轨；改用 `-m single <youtube-url>` 针对单条视频，开发调试时可结合 `--limit` 减少请求量。
- `go test ./...`：覆盖 `pkg/` 全部单元测试，是提交前的最低要求；如修改外部依赖适配层，请补充 mock，确保 CI 不依赖真实 Telegram 或 OSS。
- `docker build -t youtube-audio:latest -f Dockerfile .`：构建与生产一致的镜像，可用于部署或回归测试；若计划推送镜像，请先执行 `packer build deploy/packer/local.json` 完成内置依赖写入。

## 代码风格与命名规范
- 统一使用 Go 官方格式化工具 `gofmt`/`go fmt ./...`，保持 Tab 缩进与 <120 字符行宽；提交前可运行 `go vet ./...` 捕获常见误用。
- 对外导出类型与函数采用 PascalCase，内部辅助函数使用 lowerCamelCase；测试替身命名后缀可加 `Mock`、`Stub`，避免与真实实现混淆。
- 环境变量及配置均采用全大写下划线风格（如 `BOT_TOKEN`、`YOUTUBE_KEY`），严禁将密钥写入仓库；路径常量使用 `snake_case` 以匹配现有目录命名。

## 测试准则
- 测试文件与源码同目录，文件名以 `_test.go` 结尾，函数使用 `TestXxx` 命名；推荐表驱动写法，参考 `pkg/handler/fetcher_test.go` 和 `pkg/util/resource/resource_parser_test.go`。
- 关键集成点（Telegram Bot、OSS、阿里云存储）需增加 mock 或开关控制，避免 CI 依赖真实资源，可借鉴 `pkg/util/db/connect_test.go` 的连接包装方式。
- 新增功能必须补充失败路径与边界条件覆盖，明确断言日志或返回值，确保 CLI 既报告成功也能给出降级方案。

## 提交与 Pull Request
- 提交信息保持祈使语、语义明确，可延续项目现有中英混排风格（如“删除...”“更新...”）；一个提交聚焦单一改动并附带必要的二进制更新说明。
- PR 描述需包含变更目的、主要模块、验证步骤（示例：`go test ./...`、`docker build ...`）及任何新的环境变量，必要时列出回滚策略。
- 若调整 Terraform/Packer，请附上相关输出或截图；功能影响用户可视行为时需补充日志片段或频道截图，帮助审核者理解影响面。

## 运行环境与安全提示
- 运行 CLI 或容器前需配置 `BOT_TOKEN`、`BOT_CHAT_ID`、`CHAT_ID`、`YOUTUBE_KEY` 及阿里云凭证；Terraform/Packer 需额外设定 `ALICLOUD_ACCESS_KEY`、`ALICLOUD_SECRET_KEY` 与可选的 `ALICLOUD_REGION`。
- 大体量配置文件存放于远端 OSS（参考 README 中的 `fetch_base.json` 与 `fetch_history.json`），勿在仓库中暴露个人频道 ID 或密钥；敏感信息统一通过环境变量或私有配置仓库管理。
- 产线出现异常时，先检查 `logs/` 下的时间戳日志文件，对照 Telegram 频道或机器人消息快速比对问题批次。
