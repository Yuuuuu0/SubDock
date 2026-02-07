# SubDock - Agent Guidelines

轻量级 Web 订阅管理器，Go + Vue 3 全栈项目。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.24+, Gin, GORM, SQLite |
| 前端 | Vue 3, TypeScript, Naive UI, Pinia, Vite |
| 部署 | Docker (多阶段构建) |

---

## 构建与运行命令

### 后端 (Go)

```bash
# 构建
go build -o subdock .

# 运行 (需要先构建前端)
DATA_DIR=./data ./subdock

# 格式化
go fmt ./...

# 静态检查
go vet ./...
```

### 前端 (Vue)

```bash
cd web

# 安装依赖
pnpm install

# 开发模式 (热重载，代理到 localhost:8080)
pnpm dev

# 类型检查 + 构建
pnpm build

# 仅类型检查
npx vue-tsc --noEmit
```

### 完整构建流程

```bash
# 1. 构建前端
cd web && pnpm build && cd ..

# 2. 复制前端产物到后端
rm -rf internal/router/dist
cp -r web/dist internal/router/

# 3. 构建后端
go build -o subdock .

# 4. 运行
DATA_DIR=./data ./subdock
```

### Docker

```bash
# 构建镜像
docker build -t subdock .

# 运行
docker compose up -d
```

---

## 项目结构

```
.
├── main.go                    # 入口
├── internal/
│   ├── config/config.go       # 配置加载 (环境变量)
│   ├── model/
│   │   ├── db.go              # 数据库初始化
│   │   └── model.go           # GORM 模型定义
│   ├── handler/               # HTTP 处理器 (Gin handlers)
│   │   ├── auth.go            # 登录、改密
│   │   ├── subscription.go    # 订阅 CRUD
│   │   ├── setting.go         # 系统设置
│   │   └── config.go          # 公开配置
│   ├── middleware/auth.go     # JWT 认证中间件
│   ├── router/
│   │   ├── router.go          # 路由定义
│   │   └── static.go          # 静态文件服务
│   ├── scheduler/scheduler.go # 定时任务 (cron)
│   └── service/notifier.go    # 通知服务 (Telegram/Bark)
└── web/                       # 前端 Vue 项目
    └── src/
        ├── api/index.ts       # API 客户端 + 类型定义
        ├── router/            # Vue Router
        ├── stores/            # Pinia stores
        └── views/             # 页面组件
```

---

## 代码风格

### Go 后端

**导入顺序** (空行分隔):
```go
import (
    // 1. 标准库
    "net/http"
    "time"

    // 2. 第三方库
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    // 3. 项目内部包
    "subdock/internal/model"
)
```

**命名规范**:
- 文件名: 小写下划线 `subscription.go`
- 包名: 小写单词 `handler`, `model`
- 导出函数: PascalCase `ListSubscriptions`
- 私有函数: camelCase `formatFloat`
- 常量: PascalCase `CycleUnitMonth`

**Handler 模式**:
```go
// FunctionName 中文注释说明功能
func FunctionName(c *gin.Context) {
    // 1. 参数解析/验证
    // 2. 业务逻辑
    // 3. 返回 JSON
    c.JSON(http.StatusOK, result)
}
```

**错误处理**:
- 使用 `gin.H{"error": "中文错误信息"}` 返回错误
- HTTP 状态码: 400 参数错误, 401 未授权, 404 不存在, 500 服务器错误
- 数据库事务失败时必须 `tx.Rollback()`

**GORM 模型**:
```go
type Model struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    Name      string         `gorm:"size:128;not null" json:"name"`
    CreatedAt time.Time      `json:"created_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`  // 软删除
}
```

### TypeScript 前端

**TypeScript 配置** (严格模式):
- `strict: true`
- `noUnusedLocals: true`
- `noUnusedParameters: true`

**Vue 组件**:
- 使用 `<script setup lang="ts">` 语法
- 类型导入使用 `import type { ... }`

**API 类型定义** (在 `web/src/api/index.ts`):
```typescript
export interface Subscription {
  id?: number
  name: string
  amount: number
  // ...
}
```

**命名规范**:
- 组件文件: PascalCase `Subscriptions.vue`
- 工具文件: camelCase `index.ts`
- 接口: PascalCase `Subscription`
- 变量/函数: camelCase `fetchConfig`

---

## API 路由

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /api/config | 公开配置 | 否 |
| POST | /api/login | 登录 | 否 |
| POST | /api/change-password | 改密 | 是 |
| GET | /api/subscriptions | 订阅列表 | 是 |
| POST | /api/subscriptions | 创建订阅 | 是 |
| PUT | /api/subscriptions/:id | 更新订阅 | 是 |
| DELETE | /api/subscriptions/:id | 删除订阅 | 是 |
| POST | /api/subscriptions/:id/renew | 续订 | 是 |
| GET | /api/settings | 获取设置 | 是 |
| PUT | /api/settings | 更新设置 | 是 |

---

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DATA_DIR` | `./data` | SQLite 数据目录 |
| `PORT` | `8080` | 服务端口 |
| `JWT_SECRET` | 随机生成 | JWT 密钥 (生产必须设置) |
| `WEBSITE_TITLE` | `SubDock` | 网站标题 |
| `TZ` | `Asia/Shanghai` | 时区 |

---

## 注意事项

### 禁止

- 使用 `as any`, `@ts-ignore`, `@ts-expect-error`
- 空 catch 块 `catch(e) {}`
- 未经确认的 git commit/push
- 硬编码敏感信息 (密钥、密码)

### 必须

- Go 代码添加中文注释说明函数用途
- 数据库操作检查错误并正确处理
- 前端 API 调用处理错误状态
- 修改代码后运行 `go vet` 和 `vue-tsc` 检查

### 开发流程

1. 后端修改: 直接 `go run .` 测试
2. 前端修改: `cd web && pnpm dev` 热重载开发
3. 联调: 前端 dev server 自动代理 `/api` 到 `localhost:8080`
4. 部署: 使用 Docker 多阶段构建
