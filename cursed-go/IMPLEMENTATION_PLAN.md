# CursedChrome Go 后端重写实施计划

## 概览

将 ~3000 行 Node.js 后端（server.js + api-server.js + database.js + utils.js）全量重写为 Go，部署改为单一二进制 + 单一 Dockerfile。

### 现有 Node.js 后端清单

| 模块 | 文件 | 行数 | 关键能力 |
|------|------|------|----------|
| 主服务（cluster + WS + Proxy + 调度） | server.js | 1307 | WebSocket(4343)、HTTP Proxy(8080)、Redis 路由、RPC |
| REST API | api-server.js | 1044 | 27 路由、session 认证、bot 管理、远控 |
| Sequelize 模型 | database.js | 575 | 6 张表、关联、初始化 admin |
| 工具 | utils.js | 59 | bcrypt、随机串、彩色日志 |
| 端口 | - | - | 4343 / 8080 / 8118 |

### Go 技术栈

| 角色 | 选型 | 理由 |
|------|------|------|
| Web 框架 | **chi** | 轻量、idiomatic、与 net/http 兼容、便于嵌套路由 |
| ORM | **gorm.io/gorm** v2 + postgres driver | 与 Sequelize 概念最接近，支持自动迁移 |
| WebSocket | **nhooyr.io/websocket** (或 gorilla/websocket) | 现代、context-aware |
| Redis | **redis/go-redis** v9 | 官方推荐，pubsub 完备 |
| 配置 | **caarlos0/env** + dotenv | 12-factor 风格 |
| 日志 | **zerolog** | 高性能、结构化 |
| Session | **gorilla/sessions** + cookie store | 与 client-sessions 等价 |
| 密码 | `golang.org/x/crypto/bcrypt` | 标准库 |
| Schema 校验 | `go-playground/validator` v10 | 用 struct tag 而非 JSON schema |
| 测试 | `testing` + `stretchr/testify` + `dockertest` | 单元 + 集成 |
| HTTP 代理 | 自实现（基于 `goproxy` 或 `httputil`） | AnyProxy 在 Go 中无对应物 |

### 项目结构

```
cursed-go/
├── go.mod
├── go.sum
├── Makefile                # 构建/测试/lint 入口
├── Dockerfile              # 多阶段构建
├── docker-compose.yaml     # 替代上层 docker-compose
├── .env.example
├── cmd/
│   └── cursed-server/
│       └── main.go         # 入口
├── internal/
│   ├── config/             # 配置加载
│   ├── db/
│   │   ├── models/         # GORM 模型 (User, Bot, BotRecording, BotScreenshot, BotKeyboardLog, Setting)
│   │   ├── migrate.go      # AutoMigrate + 默认数据
│   │   └── conn.go         # 连接池
│   ├── auth/
│   │   ├── bcrypt.go
│   │   ├── session.go      # gorilla/sessions 适配
│   │   └── middleware.go
│   ├── api/
│   │   ├── server.go       # chi router 装配
│   │   ├── login.go
│   │   ├── bots.go
│   │   ├── settings.go
│   │   ├── media.go        # 截图/键盘日志/录音
│   │   ├── remote.go       # 远控
│   │   └── proxy_creds.go
│   ├── ws/
│   │   ├── server.go       # WS 4343
│   │   ├── session.go      # 单 bot 会话状态机
│   │   ├── handlers/       # PING/SYNC/SYNC_HUGE/STATE/...
│   │   ├── rpc.go          # request/response 关联表 (替代 REQUEST_TABLE)
│   │   └── auth.go         # AUTH 握手
│   ├── proxy/
│   │   ├── server.go       # HTTP Proxy 8080
│   │   ├── auth.go
│   │   └── forwarder.go    # 通过 WS RPC 转发到 bot
│   ├── busx/               # 跨进程消息总线 (Redis pubsub)
│   │   ├── client.go
│   │   ├── channels.go     # TOBROWSER_*, TOPROXY_*, SYSTEM_EVENTS
│   │   └── router.go
│   ├── messaging/          # 上线/离线/域名通知（替代 Message 类）
│   ├── utils/
│   │   ├── crypto.go       # 安全随机串
│   │   ├── logger.go
│   │   └── consts.go       # BOT_DEFAULT_*
│   └── version/
├── pkg/                    # 可被外部引用的公共包（暂留空）
├── test/
│   ├── integration/        # 黑盒集成测试
│   ├── smoke/              # smoke 脚本
│   └── fixtures/
└── scripts/
    ├── dev.sh              # 启动 dev 环境
    └── smoke.sh            # smoke 测试入口
```

## 阶段计划

每个阶段都遵循 **TDD：红 → 绿 → 重构 → 验证**，每个阶段结束必须：
1. 单元测试 100% 通过
2. lint (`golangci-lint run`) 0 warning
3. `go vet` 0 warning
4. `go build ./...` 成功
5. 提交一个 commit

### Stage 0：项目骨架
**Goal**：建立 Go 项目结构、go.mod、Dockerfile、Makefile、CI 入口。
**Success Criteria**：`go build ./...` 成功；`make test`、`make lint`、`make smoke` 命令存在。
**Tests**：版本号端点 / health check 端点。
**Status**：Complete

### Stage 1：DB 层 + Models + 迁移
**Goal**：6 张表的 GORM 模型 + AutoMigrate + 默认 admin / SESSION_SECRET 初始化。
**Success Criteria**：
- 启动后数据库出现 6 张表
- `users` 表自动出现 admin 账户（首次随机密码打印到日志）
- `settings` 表自动出现 SESSION_SECRET
- 重启不会重复创建
**Tests**：
- 模型 CRUD 单元测试（用 sqlite in-memory 或 dockertest postgres）
- 迁移幂等测试
- bcrypt 校验测试
**Status**：Complete

### Stage 2：认证 + Session 中间件
**Goal**：登录、登出、修改密码、me、CSP/Security Headers 中间件、session middleware。
**Success Criteria**：
- POST /api/v1/login 正确处理成功/失败/锁定
- 通过 session cookie 维持会话
- 修改密码后必须能用新密码登录
**Tests**：
- 登录成功/密码错误/用户不存在
- session 过期/续期
- 修改密码后旧 session 失效（如适用）
**Status**：Complete

### Stage 3：REST API (27 路由)
**Goal**：实现全部 27 条 REST 路由（不含 WS / Proxy）。
**Success Criteria**：
- 所有路由通过 contract test（请求/响应 shape 与 Node.js 对齐）
- bots 列表分页、过滤等价
- 远控接口能成功转发 RPC 到 WS（依赖 Stage 4/5）
**Tests**：
- 每条路由至少 1 个 happy path + 1 个错误路径
- 静态资源托管 + CSP header 校验
**Status**：Complete

### Stage 4：Redis Pub/Sub + Bot 路由总线
**Goal**：跨进程/单进程统一的消息总线，支持 `TOBROWSER_*`、`TOPROXY_*`、`SYSTEM_EVENTS` 三类 channel。
**Success Criteria**：
- 多副本部署下，bot 路由仍正确
- pubsub 重连容错
**Tests**：
- 收发对称性
- 断线重连
- 消息序列化
**Status**：Complete

### Stage 5：WebSocket 服务 + RPC 调用表
**Goal**：实现 4343 端口 WS、AUTH 握手、PING、SYNC、SYNC_HUGE、STATE、REALTIME_IMG、SCREEN_CAPTURE_DATA、USER_ACTIVITY、DEBUG_LOG、KEYBOARD_LOGS、AUDIO_DATA 全部 RPC handler，REQUEST_TABLE 替换为 `sync.Map + context timeout`。
**Success Criteria**：
- 用现有 Chrome 扩展直连成功
- 所有 RPC action 落库正确
- timeout 后正确清理
**Tests**：
- 每个 RPC handler 单测（mock 数据库）
- WS 集成测试（用真实 client 跑握手 + ping）
- bot 在线/离线状态切换
**Status**：Complete

### Stage 6：HTTP 代理服务
**Goal**：8080 代理：Basic Auth → 通过 bot 转发请求 → 将响应回放给客户端。支持 HTTP CONNECT。
**Success Criteria**：
- curl 通过本机代理访问 example.com 成功
- 代理凭证错误返回 407
- 代理凭证正确但 bot 离线返回 502
**Tests**：
- 代理认证表
- forward 路径单测
- E2E：本机起一个 fake bot WS，用 curl 验证端到端
**Status**：Complete

### Stage 7：集成测试 + Smoke 测试 + 性能基线
**Goal**：
- `make smoke` 一键跑完：启动→登录→列 bot→开关全局代理→停服
- 对比 Node.js 后端的关键性能指标（启动时间、内存、QPS）
**Success Criteria**：smoke 全绿；性能不劣于 Node.js 版本。
**Tests**：
- bash + curl smoke 脚本
- ws-cli 模拟 bot
**Status**：Complete

### Stage 8：Dockerfile + docker-compose + 上线
**Goal**：单二进制 Dockerfile，多阶段构建，最终镜像 < 30 MB。
**Success Criteria**：
- `docker compose up` 一键起 db + redis + cursed-go
- 现有 Chrome 扩展无修改即可连上
- 在 Portainer 上替换 cursed stack，所有功能验证
**Tests**：
- 镜像启动时间 < 5s
- 容器健康检查通过
**Status**：Complete

## 关键迁移决策

1. **不再使用 cluster**：Go 天然多 goroutine，单进程足够。Redis pubsub 仍保留以支持横向扩展到多副本。
2. **Sequelize 迁移**：GORM 自动迁移负责 schema；遗留旧库切换时手工 export → import。
3. **AnyProxy 替代**：用 `goproxy` (https://github.com/elazarl/goproxy) 或自实现 ReverseProxy + CONNECT 隧道，HTTPS 用同样的 CA。
4. **JS 混淆扩展注入**：保留现有 Node 脚本（在容器构建期跑一次即可），不放进 Go 二进制。
5. **保留兼容**：所有 RPC action 名、消息格式、API URL、错误码与 Node 版完全一致 → 现有 Chrome 扩展、GUI 不需要修改。

## 验收标准

- [ ] 所有阶段单元测试 100% 通过
- [ ] 集成测试 100% 通过
- [ ] smoke 脚本绿
- [ ] 现有 Chrome 扩展直连无需改动
- [ ] 现有 Vue GUI 直连无需改动
- [ ] 容器镜像 < 30 MB
- [ ] 启动时间 < 5s
- [ ] CPU/内存 ≤ Node.js 版本
- [ ] CI 通过 lint + vet + test
