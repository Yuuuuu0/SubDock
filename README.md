# SubDock

SubDock 是一个轻量级、自托管的 Web 订阅管理器，用于记录周期性服务、跟踪到期时间，并通过 Telegram 或 Bark 发送提醒。

后端使用 Go、Gin、GORM 与 SQLite，前端使用 Vue 3、TypeScript、Naive UI 与 Pinia。生产构建会将前端静态资源嵌入 Go 二进制，部署时只需要一个容器和一个持久化数据目录。

## 功能概览

- 订阅新增、编辑、删除与搜索
- 概览页展示订阅数量、30 天内到期、已过期和正常状态
- 按币种分别估算月均费用，不对不同币种做隐式换算
- 近期到期时间线与高费用订阅提示
- 按到期日期排序并显示剩余天数和状态
- 支持免费订阅及多种货币
- 支持按天、月、季度、半年和年设置周期
- 自动计算到期日期
- 手动续订与自动续订
- 每项订阅可单独配置提前提醒天数
- Telegram 与 Bark 通知渠道
- 可配置每天的通知时段并发送测试通知
- 管理员登录、密码修改与 JWT 鉴权
- 通过 `WEBSITE_TITLE` 自定义站点名称

当前产品面向单管理员、自托管场景，不包含多用户协作、账单支付或附件管理。

## 界面与使用流程

SubDock 的主要使用流程分为四个部分：

1. 登录：首次启动后使用日志中生成的管理员密码登录。
2. 概览：查看关键指标、近期到期项目和按币种拆分的费用估算。
3. 订阅管理：维护订阅价格、周期、到期日、续订方式和提醒策略。
4. 系统设置：配置通知时段、Telegram、Bark，并修改管理员密码。

## 技术栈

| 层级 | 技术 |
|---|---|
| 后端 | Go 1.24+、Gin、GORM、SQLite |
| 前端 | Vue 3、TypeScript、Naive UI、Pinia、Vite |
| 包管理 | pnpm 10.26.0 |
| 调度 | robfig/cron |
| 部署 | Docker 多阶段构建、单二进制运行 |

## 快速开始

### Docker Compose

推荐先由当前用户创建数据目录，再启动服务。容器内应用以非 root 用户运行，其 UID/GID 为 `1000:1000`。

```bash
mkdir -p data
docker compose up -d
```

默认仅监听本机：

```text
http://127.0.0.1:8080
```

查看启动日志和首次登录密码：

```bash
docker compose logs subdock
```

检查运行状态：

```bash
docker compose ps
curl -fsS http://127.0.0.1:8080/api/config
docker inspect --format '{{json .State.Health}}' subdock
```

`docker-compose.yml` 默认拉取 `yuuuuu0/subdock:latest`。正式环境建议改用明确版本标签，避免 `latest` 带来不可预期的升级。

### 直接运行 Docker 镜像

注意数据目录必须挂载到容器内的 `/data`：

```bash
mkdir -p data

docker run -d \
  --name subdock \
  -p 127.0.0.1:8080:8080 \
  -v "$(pwd)/data:/data" \
  -e TZ=Asia/Shanghai \
  -e WEBSITE_TITLE=SubDock \
  --restart always \
  yuuuuu0/subdock:latest
```

如果 Linux 主机上的数据目录不是 UID/GID `1000:1000` 且出现权限错误，可调整目录所有者：

```bash
sudo chown -R 1000:1000 data
```

## 首次登录

- 管理员用户名固定为 `admin`。
- 首次创建数据库时，应用会生成随机初始密码并输出到日志。
- 登录后应立即修改密码。

Docker Compose：

```bash
docker compose logs subdock | grep "密码"
```

直接运行容器：

```bash
docker logs subdock | grep "密码"
```

## 数据目录与 JWT 密钥

容器内 `DATA_DIR` 默认为 `/data`，本地直接运行时默认为 `./data`。目录中包含：

```text
data/
├── subdock.db   # SQLite 主数据库
└── .jwt-secret  # 未显式配置 JWT_SECRET 时自动生成的签名密钥
```

JWT 密钥采用以下规则：

1. 如果设置了非空 `JWT_SECRET`，应用使用环境变量中的值。
2. 如果未设置，应用生成安全随机密钥并保存到 `DATA_DIR/.jwt-secret`。
3. 后续启动会复用该文件，避免每次重启都让现有登录状态失效。
4. 自动生成的密钥文件权限为 `0600`。
5. 手动设置的 `JWT_SECRET` 至少需要 32 个字符，且不能使用示例或常见弱密钥。

`.jwt-secret` 与数据库同样重要，备份和恢复时必须一起处理。删除或替换它不会损坏订阅数据，但会让之前签发的登录令牌失效。

如需自行管理密钥，可在 Compose 同目录创建 `.env`：

```dotenv
JWT_SECRET=请替换为至少32字符的高强度随机字符串
```

不要将包含真实密钥的 `.env` 提交到版本库。

## 环境变量

| 变量 | 本地默认值 | 容器默认值 | 说明 |
|---|---|---|---|
| `DATA_DIR` | `./data` | `/data` | SQLite 数据库和持久化 JWT 密钥目录 |
| `PORT` | `8080` | `8080` | HTTP 服务端口，范围为 1-65535 |
| `JWT_SECRET` | 自动生成并持久化 | 自动生成并持久化 | 可选的 JWT 签名密钥，手动设置时至少 32 字符 |
| `WEBSITE_TITLE` | `SubDock` | `SubDock` | 浏览器标题和站点名称 |
| `TZ` | 由操作系统决定 | `Asia/Shanghai` | 进程时区，影响提醒调度和日期判断 |

## 通知配置

### Telegram

1. 通过 `@BotFather` 创建机器人并取得 Bot Token。
2. 通过 `@userinfobot` 等方式取得 Chat ID。
3. 在系统设置页填写配置。
4. 保存后使用“测试”按钮验证连通性。

### Bark

1. 在 iOS 设备安装 Bark。
2. 获取 Bark 推送地址，例如 `https://api.day.app/<your-key>`。
3. 在系统设置页填写地址并发送测试通知。

通知凭据会保存在 SQLite 数据库中，请保护数据目录和备份文件，不要将其公开上传。

## 本地开发

### 环境要求

- Go 1.24 或更高版本
- Node.js 20 或更高版本
- Corepack
- pnpm 10.26.0，由 `web/package.json` 的 `packageManager` 字段固定

启用 Corepack：

```bash
corepack enable
```

### 启动后端

```bash
DATA_DIR=./data go run .
```

仓库会跟踪 `internal/router/dist/.gitkeep`，因此干净检出后可以直接编译和运行后端。开发前端时，页面由 Vite 提供，不依赖嵌入的生产构建产物。

### 启动前端

另开一个终端：

```bash
cd web
pnpm install --frozen-lockfile
pnpm dev
```

Vite 开发服务器会把 `/api` 请求代理到 `http://localhost:8080`。

### 类型检查与生产构建

```bash
cd web
pnpm typecheck
pnpm build
```

### 完整本地构建

仓库根目录提供统一构建脚本。脚本会严格按照 pnpm 锁文件安装依赖、构建前端、复制嵌入资源并编译 Go 二进制：

```bash
./build.sh
```

构建结果为仓库根目录下的 `subdock`：

```bash
DATA_DIR=./data ./subdock
```

## 自行构建 Docker 镜像

```bash
docker build -t subdock:local .

docker run --rm \
  -p 127.0.0.1:8080:8080 \
  -v "$(pwd)/data:/data" \
  subdock:local
```

镜像采用三阶段构建：

1. Node.js + pnpm 构建前端。
2. Go + CGO 构建包含前端资源的后端二进制。
3. Alpine 运行镜像以非 root 用户启动，并通过 `/api/config` 执行健康检查。

## 备份

### 推荐方式：停机复制整个数据目录

这是最简单可靠的方式，并会同时备份数据库和 JWT 密钥：

```bash
backup_dir="backups/subdock-$(date +%Y%m%d-%H%M%S)"

docker compose stop subdock
mkdir -p "$backup_dir"
cp -a data/. "$backup_dir/"
docker compose start subdock
```

确认备份至少包含：

```text
subdock.db
.jwt-secret
```

如果通过环境变量管理 `JWT_SECRET`，还需要使用安全的密钥管理方式单独备份该值。

### 使用 SQLite 在线备份

主机安装了 `sqlite3` 时，可以使用 SQLite 的备份命令：

```bash
mkdir -p backups
sqlite3 data/subdock.db ".backup 'backups/subdock.db'"
cp -a data/.jwt-secret backups/.jwt-secret
```

备份后可以执行完整性检查：

```bash
sqlite3 backups/subdock.db 'PRAGMA integrity_check;'
```

## 恢复

恢复前停止服务，并保留当前数据目录作为额外保险：

```bash
docker compose stop subdock
mv data "data.before-restore-$(date +%Y%m%d-%H%M%S)"
mkdir -p data
cp -a backups/<备份目录>/. data/
sudo chown -R 1000:1000 data
docker compose up -d
```

恢复后检查：

```bash
docker compose ps
docker compose logs --tail=100 subdock
curl -fsS http://127.0.0.1:8080/api/config
```

## 升级

升级前务必先备份整个数据目录：

```bash
backup_dir="backups/subdock-before-upgrade-$(date +%Y%m%d-%H%M%S)"
docker compose stop subdock
mkdir -p "$backup_dir"
cp -a data/. "$backup_dir/"
docker compose start subdock
```

然后拉取并启动新镜像：

```bash
docker compose pull
docker compose up -d
docker compose ps
docker compose logs --tail=100 subdock
```

如果升级后异常，停止服务并按“恢复”章节还原升级前备份，再切回原镜像版本。

## 数据兼容性

本轮工程化、目录和界面重构不修改现有业务数据表结构，已有 `subdock.db` 可以直接复用，不需要手工迁移。

仍建议在每次升级前备份数据库和 `.jwt-secret`。未来如果版本需要调整数据结构，发布说明会给出迁移和回滚步骤。

## 验证命令

提交变更或制作镜像前建议执行：

```bash
# Shell 与 Compose 配置
bash -n build.sh
docker compose config

# Go
go test ./...
go vet ./...

# 前端
cd web
pnpm install --frozen-lockfile
pnpm typecheck
pnpm build
cd ..

# 容器
docker build -t subdock:verify .
```

运行中的服务可以使用以下命令进行健康检查：

```bash
curl -fsS http://127.0.0.1:8080/api/config
docker inspect --format '{{json .State.Health}}' subdock
```

## API 概览

| 方法 | 路径 | 说明 | 认证 |
|---|---|---|---|
| `GET` | `/api/config` | 获取公开站点配置 | 否 |
| `POST` | `/api/login` | 管理员登录 | 否 |
| `POST` | `/api/change-password` | 修改管理员密码 | 是 |
| `GET` | `/api/subscriptions` | 获取订阅列表 | 是 |
| `POST` | `/api/subscriptions` | 创建订阅 | 是 |
| `PUT` | `/api/subscriptions/:id` | 更新订阅 | 是 |
| `DELETE` | `/api/subscriptions/:id` | 删除订阅 | 是 |
| `POST` | `/api/subscriptions/:id/renew` | 手动续订 | 是 |
| `POST` | `/api/subscriptions/:id/test-notify` | 测试单项订阅通知 | 是 |
| `GET` | `/api/settings` | 获取系统设置 | 是 |
| `PUT` | `/api/settings` | 更新系统设置 | 是 |
| `POST` | `/api/settings/test-notify` | 测试通知渠道 | 是 |

## 安全建议

- 默认 Compose 只绑定 `127.0.0.1`；需要公网访问时，请通过带 HTTPS 的反向代理暴露服务。
- 首次登录后立即修改管理员密码。
- 不要使用示例 JWT 密钥，也不要把真实密钥提交到 Git。
- 限制 `data/` 与备份目录的文件权限。
- 不要公开包含初始密码、Bot Token 或 Bark 地址的日志与截图。
- 定期备份并验证备份数据库的完整性。

## License

MIT
