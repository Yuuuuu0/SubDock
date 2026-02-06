# SubDock

轻量级 Web 订阅管理器，支持按到期时间追踪订阅并通过 Telegram/Bark 发送提醒。

## 主要功能

- 订阅管理：新增、编辑、删除、搜索
- 自动计算到期日：由开始日期 + 周期自动计算
- 默认排序：按到期时间升序（最近到期排最前）
- 免费订阅支持：金额可为 `0`
- 提醒策略：可配置提前 N 天提醒
- 通知渠道：Telegram、Bark
- 通知时段：可配置每天具体发送小时（0-23）
- 网站标题可配置：支持 `WEBSITE_TITLE`

## 技术栈

- 后端：Go、Gin、GORM、SQLite
- 前端：Vue 3、TypeScript、Naive UI、Pinia
- 打包部署：Docker（多阶段构建）

## 快速开始

### 1) Docker Compose（推荐）

```bash
docker compose up -d --build
```

默认访问地址：`http://localhost:8080`

### 2) 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `DATA_DIR` | `./data` | 数据目录（SQLite 文件所在目录） |
| `PORT` | `8080` | 服务监听端口 |
| `JWT_SECRET` | 随机生成 | JWT 密钥，生产环境必须显式设置 |
| `WEBSITE_TITLE` | `SubDock` | 前端显示的网站标题 |
| `TZ` | `Asia/Shanghai` | 容器时区 |

### 3) 首次登录

- 用户名固定为：`admin`
- 初始密码会在服务首次启动时输出到日志

```bash
docker logs subdock | grep "初始密码"
```

## 本地开发

### 前端

```bash
cd web
pnpm install
pnpm build
```

### 后端

前端构建后，将 `web/dist` 同步到 `internal/router/dist` 再编译：

```bash
cd ..
rm -rf internal/router/dist
cp -r web/dist internal/router/
go build -o subdock .
DATA_DIR=./data ./subdock
```

## 通知配置说明

### Telegram

1. 使用 `@BotFather` 创建机器人，拿到 `Bot Token`
2. 使用 `@userinfobot` 获取你的 `Chat ID`
3. 在系统设置页填写并测试

### Bark

1. 安装 Bark（iOS）
2. 获取 Bark URL（如：`https://api.day.app/<your-key>`）
3. 在系统设置页填写并测试

## License

MIT
