# 开发交接文档

最后更新：2026-03-07

## 1. 这份文档的目的

这份文档面向下一位接手开发的工程师，目标是快速说明：

- 当前代码和任务处于什么状态
- 应该在哪个 worktree / 分支继续工作
- 哪些运行方式是真实链路，哪些只是开发期内存桩
- 已知的高风险点、容易误判的问题、重要环境配置
- 当前远端分支和本地分支的真实关系

## 2. 当前工作区与 Git 状态

本次后端开发一直在 backend worktree 中进行：

- backend worktree 路径：`/Users/tylor/Code/GoAIPj-backend`
- 当前本地分支：`backend`

当前几个关键 commit / 远端状态：

- 本地 `backend` HEAD：`b8b382d merge: integrate origin/codex/feature into backend`
- 远端 `origin/backend`：`4d0ea62 fix: repair historical resource graph state`
- 远端 `origin/codex/feature`：`b8b382d merge: integrate origin/codex/feature into backend`

这意味着：

1. 本地 `backend` 分支已经包含前后端融合后的 merge commit。
2. 这个 merge commit 已经推送到了远端 `codex/feature`。
3. 这个 merge commit **没有** 推回 `origin/backend`。

对下一位开发者最重要的含义：

- 如果你要继续“纯后端线”开发，请先明确你要基于：
  - 当前本地 `backend`（已经融合了 frontend）
  - 还是远端 `origin/backend`（仍然是纯后端线）
- 不要在没有想清楚分支策略前，直接把本地 `backend` 推回 `origin/backend`。

## 3. 任务完成状态

当前 [backend_tasklist.md](/Users/tylor/Code/GoAIPj-backend/docs/backend_tasklist.md) 中：

- `BE-001` 到 `BE-071` 全部已标记 `已完成`
- `MVP` 全部完成
- `Post-MVP` 全部完成

## 4. 关键文档

建议先读这些文档：

- [backend_design.md](/Users/tylor/Code/GoAIPj-backend/docs/backend_design.md)
- [backend_tasklist.md](/Users/tylor/Code/GoAIPj-backend/docs/backend_tasklist.md)
- [techstack.md](/Users/tylor/Code/GoAIPj-backend/docs/techstack.md)
- [requirement.md](/Users/tylor/Code/GoAIPj-backend/docs/requirement.md)
- [frontend_api.md](/Users/tylor/Code/GoAIPj-backend/docs/frontend_api.md)

## 5. 当前后端运行方式

### 5.1 API 启动方式

开发期 API 入口：

```bash
go run ./apps/api/cmd/api
```

入口文件：

- [main.go](/Users/tylor/Code/GoAIPj-backend/apps/api/cmd/api/main.go)
- [bootstrap.go](/Users/tylor/Code/GoAIPj-backend/apps/api/cmd/api/bootstrap.go)

非常重要：

- 当前 `cmd/api` 默认启动的是**内存态开发服务器**
- `group/resource/graph/chat/job/event/storage` 都是 in-memory / fake 实现
- 重启后数据会清空
- 如果前端联调看到“代码说已经修了，但接口行为还像旧版本”，第一怀疑对象不是代码，而是 **8080 上跑着旧进程**

此前真实遇到过一次：

- 前端复验失败
- 根因是 `127.0.0.1:8080` 上仍然挂着更早启动的旧 `go run` 进程
- 通过停掉旧进程并重启最新 backend worktree 后恢复正常

### 5.2 Worker 启动方式

Worker 入口：

```bash
go run ./apps/api/cmd/worker
```

文件：

- [main.go](/Users/tylor/Code/GoAIPj-backend/apps/api/cmd/worker/main.go)

重要说明：

- 当前 `worker` 入口只做了基础 `asynq` mux 注册
- 它并没有像 API 一样把完整业务依赖全部 wiring 到真实 handler
- 所以它更接近“入口占位 + 基础注册”，不是完整生产 worker

## 6. 本地依赖与真实基础设施

项目已经提供：

- [docker-compose.yml](/Users/tylor/Code/GoAIPj-backend/docker-compose.yml)

包含：

- PostgreSQL `5432`
- Redis `6379`
- MinIO `9000/9001`
- Tika `9998`

但要注意：

- `docker-compose` 提供的是“真实外部依赖”
- 当前 `cmd/api` 并没有默认接入这些依赖，而是走内存态 bootstrap
- 如果下一位开发者要推进真实持久化 / 真 worker / 真实任务链路，需要继续把 API/Worker 从 in-memory bootstrap 迁到真实 infra wiring

## 7. 关键配置与工具

### 7.1 配置系统

配置包：

- [config.go](/Users/tylor/Code/GoAIPj-backend/internal/config/config.go)

当前启动会读取：

- `.env`
- 或环境变量

### 7.2 sqlc

`sqlc` 配置：

- [sqlc.yaml](/Users/tylor/Code/GoAIPj-backend/sqlc.yaml)

本机上 `sqlc` 曾经不在 PATH，实际可用路径是：

```bash
/Users/tylor/Code/GoAIPj/bin/sqlc
```

生成命令建议用：

```bash
env PATH=/Users/tylor/Code/GoAIPj/bin:$PATH sqlc generate
```

### 7.3 真实 AI 集成

真实百炼适配层在：

- [aliyun.go](/Users/tylor/Code/GoAIPj-backend/internal/infra/llm/aliyun.go)

之前真实集成测试使用过：

```bash
env ALIYUN_BAILIAN_API_KEY_FILE=/Users/tylor/Code/GoAIPj/docs/api.md GOCACHE=/tmp/goaipj-backend-cache GOMODCACHE=/tmp/goaipj-backend-gomodcache go test ./internal/http/api -run TestConversationEndpointsWithAliyunBailian -count=1 -v -timeout 90s
```

注意：

- 这个环境变量命名很怪，指向的是一个本机文件路径
- 接手人不要假设它是标准 secret 管理方式
- 如果这个文件不存在，真实 AI 测试会失败

## 8. 当前测试方式

### 8.1 后端全量测试

推荐命令：

```bash
env GOCACHE=/tmp/goaipj-backend-cache GOMODCACHE=/tmp/goaipj-backend-gomodcache go test ./...
```

### 8.2 前端校验

由于当前本地 `backend` 分支已经融合 frontend 代码，下面这些前端命令在这个 worktree 里也是可跑的：

```bash
npm ci
npm run build
npm test
npm run lint
```

上述命令在最近一次 merge 后已实际跑通。

## 9. 当前实现里最重要的事实

### 9.1 API 统一响应包装

统一 envelope：

```json
{
  "data": ...,
  "error": ...,
  "meta": ...
}
```

主要代码在：

- [server.go](/Users/tylor/Code/GoAIPj-backend/internal/http/api/server.go)

### 9.2 fresh resource graph 物化

当前 `GET /api/v1/resources/{resourceId}/graph` 在 graph 缺失时会做 on-demand materialization。

相关代码：

- [server.go](/Users/tylor/Code/GoAIPj-backend/internal/http/api/server.go)

边界说明：

- 这是开发/联调友好的最小图谱，不等于完整异步高质量图谱
- 主要目的是打通“新资源 -> 图谱页面”的闭环

### 9.3 历史状态不一致自修复

之前出现过真实问题：

- 资源 `status=uploaded`
- 分组 `completed_resource_count=0`
- 但框架图谱已经 `200`

修复方式：

- 在读取 `group detail / group resources / resource detail` 时，如果资源已有 active resource graph，但状态还没回到 `completed`，后端会做读时修复

相关代码：

- [server.go](/Users/tylor/Code/GoAIPj-backend/internal/http/api/server.go)
- [groups_handler_test.go](/Users/tylor/Code/GoAIPj-backend/internal/http/api/groups_handler_test.go)

重要提醒：

- 这类修复只会在**新代码进程**里生效
- 如果前端还看到旧行为，先确认 8080 是否是最新进程

### 9.4 分组删除级联清理

当前 group 删除会级联清理：

- resources
- resource graphs
- framework graphs
- conversations/messages
- jobs
- 资源关联对象存储对象

相关代码：

- [service.go](/Users/tylor/Code/GoAIPj-backend/internal/app/group/service.go)
- [group_cleanup.go](/Users/tylor/Code/GoAIPj-backend/internal/http/api/group_cleanup.go)

### 9.5 可观测接入

Post-MVP 已增加：

- Tracer（默认 noop）
- ErrorReporter（默认 noop）

包位置：

- [observability.go](/Users/tylor/Code/GoAIPj-backend/internal/infra/observability/observability.go)

当前已接入链路：

- `http.request`
- 资源上传
- 框架图谱生成
- 分组上下文检索
- 问答生成
- SSE 消息流
- 分组事件发布
- 5xx 错误聚合上报

这意味着：

- 目前没有强依赖真实 OpenTelemetry / Sentry
- 但已经有抽象层，接真实 provider 时不需要重写业务逻辑

## 10. communication 协作规则

之前前后端协作使用了：

- `Communication/communication.md`

规则是：

1. 前后端轮流写 `communication time count N`
2. 超过 `10` 条后，必须先压缩为：
   - `front_end_privious_summary:`
   - `back_end_privious_summary:`
3. 然后重新从 `communication time count 1` 开始

最近一次状态：

- communication 已经重新压缩过
- 当前最新一轮的编号曾重置到 `count 1`

接手人如果继续写 communication，先确认当前文件中最后一个编号，不要复用旧编号。

## 11. 已知坑点

### 11.1 “行为没修好”不一定是代码问题

优先检查：

- `127.0.0.1:8080` 是不是旧进程
- 是否重启过 API
- dev 服务是否因为是内存态导致旧样本已经不存在

### 11.2 dev API 是内存态

后果：

- 重启会丢数据
- 前端如果拿某个固定 `group_id/resource_id` 反复复验，重启后很可能直接 `404`
- 不适合拿它当“可持久化环境”

### 11.3 本地 backend 分支和远端 backend 分支已不一致

当前：

- 本地 `backend` 包含 merge commit `b8b382d`
- 远端 `origin/backend` 仍停在 `4d0ea62`
- 远端 `origin/codex/feature` 已是 `b8b382d`

所以接手人第一件事应该是先决定：

- 继续把 `backend` 当成后端线
- 还是以后统一基于 `codex/feature`

### 11.4 Worker 不是完整真链路

目前 worker 入口偏骨架态，不要假设：

- 启动 worker == 完整任务链路已接通

如果要继续推进真实异步任务，需要补依赖 wiring。

## 12. 推荐接手顺序

建议下一位开发者按这个顺序进入：

1. 先确认你要基于哪个分支继续：
   - `origin/backend`
   - 还是 `origin/codex/feature`
2. 启动 API 前先确认 8080 没有旧进程。
3. 如果需要真实联调，不要依赖旧样本 ID，直接新建样本复验。
4. 如果要推进生产化，优先做：
   - API 从 in-memory bootstrap 迁到真实 infra
   - Worker 真正 wiring
   - 持久化 / 队列 / 对象存储 / Tika 真链路

## 13. 这次交接时的最终结论

截至本次交接：

- backend 任务清单已全部完成
- 本地 `backend` 分支已融合 frontend
- 融合结果已推到 `origin/codex/feature`
- `origin/backend` 尚未同步这个融合结果

如果下一位开发者不知道从哪开始，先从“分支策略确认”和“API 是否为最新进程”这两个问题入手，能避免大部分无效排查。
