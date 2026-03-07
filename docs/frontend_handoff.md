# 前端交接文档

## 1. 当前结论

前端 `MVP` 与 `Post-MVP` 开发已经完成，当前代码默认直连真实后端，不再默认使用 mock。

当前前端已完成并验证的范围：

- 分组列表：创建、重命名、删除、最近访问、状态统计
- 分组详情：资源列表、网页资源创建、文件上传入口、框架图谱预览
- 资源图谱页：图谱浏览、节点详情、扩展节点、问答、流式回复
- 分组框架图谱页：读取、筛选、导出 SVG
- UI 能力：浅/深色模式、全局反馈、面包屑、空态/错误态、会话恢复
- 联调能力：真实 REST API、SSE、流式聊天 SSE、live smoke test

当前交接时的判断：

- 前端主链路已可验收
- 默认开发模式是 `live`
- mock 仍保留，但只用于显式切换

## 2. 关键文档

接手前先看这些文档：

- [front_requirement.md](/Users/tylor/Code/GoAIPj/docs/front_requirement.md)
- [frontend_api.md](/Users/tylor/Code/GoAIPj/docs/frontend_api.md)
- [frontend_tasklist.md](/Users/tylor/Code/GoAIPj/docs/frontend_tasklist.md)
- [techstack.md](/Users/tylor/Code/GoAIPj/docs/techstack.md)
- [communication.md](/Users/tylor/Code/GoAIPj/Communication/communication.md)

设计稿参考：

- [design/index.html](/Users/tylor/Code/GoAIPj/design/index.html)
- [design/styles.css](/Users/tylor/Code/GoAIPj/design/styles.css)

## 3. 当前分支与 Git 状态

交接时本地当前分支是：

- `codex/feature`

远程已存在：

- `origin/codex/feature`

前端最近关键提交：

- `5270796` `fix(frontend): default to live api flows`
- `fded13f` `feat(frontend): persist graph conversations`
- `c8fd7c9` `feat(frontend): add richer graph filters`
- `22374ec` `feat(frontend): export graphs as svg`
- `62ce49d` `feat(frontend): add post-mvp dark mode`
- `6a8fe3d` `test(frontend): add live acceptance smoke test`
- `1c22375` `fix(frontend): support live stream event aliases`
- `74c1c08` `feat(frontend): adapt live api payloads`

注意：

- 工作区里存在一些非前端开发产生的未提交改动和未跟踪文件，比如 `.DS_Store`、`docs/backend_tasklist.md`、`package.json`、`docker-compose.yml`、`migrations/...`
- 这些不一定属于前端接手范围，提交前要先确认归属
- 不要直接清空工作区

## 4. 运行方式

### 4.1 前端

项目脚本定义在 [package.json](/Users/tylor/Code/GoAIPj/package.json)。

常用命令：

```bash
npm run dev
npm test
npm run lint
npm run build
```

默认开发服务：

- Vite 默认端口：`5173`
- 本地验收期间也曾有一个前端服务跑在 `4174`

### 4.2 当前 live / mock 机制

关键文件：

- [src/main.tsx](/Users/tylor/Code/GoAIPj/src/main.tsx)
- [vite.config.ts](/Users/tylor/Code/GoAIPj/vite.config.ts)

当前规则：

- 默认是 `live`
- 只有设置 `VITE_API_MODE=mock` 才启动 MSW
- 默认代理目标是 `http://127.0.0.1:8080`
- 可通过 `VITE_LIVE_API_TARGET` 覆盖 live 后端地址

行为说明：

- `VITE_API_MODE=mock`：启动 MSW，使用本地 mock
- 未设置 `VITE_API_MODE`：默认真实联调

## 5. 后端联调前提

当前前端默认依赖真实后端：

- `http://127.0.0.1:8080`

最基本检查：

```bash
curl -sf http://127.0.0.1:8080/readyz
curl -sf http://127.0.0.1:8080/api/v1/groups
```

重要背景：

- 后端 dev 服务是内存态
- 后端重启后，之前创建的 group/resource 样本会消失
- 所以联调时不要依赖历史 id 长期存在

## 6. 已完成的真实联调结论

基于真实后端，当前已验证通过的链路包括：

- `GET /readyz`
- 分组列表 / 分组详情
- 创建分组
- 创建网页资源
- 资源图谱读取
- 分组框架图谱读取
- 节点详情读取
- 创建会话
- 非流式消息
- 流式消息 SSE
- 分组事件 SSE

最近一次真实问题已经确认解决：

- 之前出现“资源图谱已可用，但资源状态仍是 `uploaded`、`completed_resource_count` 仍是 `0`”的问题
- 根因不是前端误判，而是 `127.0.0.1:8080` 当时跑的是旧 backend dev 进程
- 后端重启到最新代码后，新样本已验证：
  - 资源状态会变为 `completed`
  - 分组 `completed_resource_count` 会正确变为 `1`

相关沟通记录在：

- [communication.md](/Users/tylor/Code/GoAIPj/Communication/communication.md)

## 7. 关键验证脚本

真实 smoke test：

- [liveSmoke.mjs](/Users/tylor/Code/GoAIPj/src/test/liveSmoke.mjs)

运行方式：

```bash
node src/test/liveSmoke.mjs
```

默认请求：

- `http://127.0.0.1:8080/api/v1`

它覆盖的链路：

- readyz
- create group
- create web resource
- get group detail
- get resources
- get resource graph
- create conversation
- non-stream message
- stream message SSE
- group events SSE connect

建议：

- 每次后端重启或联调进程切换后，先跑一次这个脚本

## 8. 目录与职责

前端主要目录：

- [src/api](/Users/tylor/Code/GoAIPj/src/api)：请求封装、类型、live adapter
- [src/app](/Users/tylor/Code/GoAIPj/src/app)：应用入口与 provider
- [src/components](/Users/tylor/Code/GoAIPj/src/components)：布局、图谱、基础 UI、主题、反馈
- [src/hooks](/Users/tylor/Code/GoAIPj/src/hooks)：Query hooks、SSE、会话恢复、最近访问等
- [src/mocks](/Users/tylor/Code/GoAIPj/src/mocks)：MSW mock
- [src/pages](/Users/tylor/Code/GoAIPj/src/pages)：页面组件
- [src/store](/Users/tylor/Code/GoAIPj/src/store)：Zustand 图谱工作台状态
- [src/styles](/Users/tylor/Code/GoAIPj/src/styles)：全局 token 与基础样式
- [src/test](/Users/tylor/Code/GoAIPj/src/test)：测试 setup 与 live smoke

关键实现文件：

- [src/api/liveAdapters.ts](/Users/tylor/Code/GoAIPj/src/api/liveAdapters.ts)
- [src/api/conversations.ts](/Users/tylor/Code/GoAIPj/src/api/conversations.ts)
- [src/hooks/useGroupEvents.ts](/Users/tylor/Code/GoAIPj/src/hooks/useGroupEvents.ts)
- [src/pages/GroupsPage.tsx](/Users/tylor/Code/GoAIPj/src/pages/GroupsPage.tsx)
- [src/pages/GroupDetailPage.tsx](/Users/tylor/Code/GoAIPj/src/pages/GroupDetailPage.tsx)
- [src/pages/FrameworkGraphPage.tsx](/Users/tylor/Code/GoAIPj/src/pages/FrameworkGraphPage.tsx)
- [src/pages/ResourceGraphPage.tsx](/Users/tylor/Code/GoAIPj/src/pages/ResourceGraphPage.tsx)
- [src/components/graph/KnowledgeGraph.tsx](/Users/tylor/Code/GoAIPj/src/components/graph/KnowledgeGraph.tsx)

## 9. 这份代码里最关键的“坑”

### 9.1 默认不是 mock

这是最重要的一点。

现在默认是 live，不是 mock。很多“页面没数据”或“请求 404”的第一反应，不该再去怀疑 mock 没起，而应该先检查：

- 后端进程是否真的跑在 `127.0.0.1:8080`
- 当前进程是不是最新代码
- 后端内存态数据是不是被重启清空了

### 9.2 后端存在字段别名与契约漂移

前端已经在 adapter 层兼容过这些差异，不建议页面层直接重新适配。

已兼容的例子包括：

- `data: []` 数组结果转成前端分页结构
- graph edge 的 `source_id / target_id / relation`
- message 的 `cited_chunk_ids / cited_node_ids / context_snapshot`
- SSE 事件名 `assistant.delta / assistant.message / done`

如果后端再变字段，优先改 [liveAdapters.ts](/Users/tylor/Code/GoAIPj/src/api/liveAdapters.ts)，不要把兼容逻辑撒到页面里。

### 9.3 图谱重复节点边问题

之前真实数据出现过重复节点/边，导致 React duplicate key warning。

当前已经在 [liveAdapters.ts](/Users/tylor/Code/GoAIPj/src/api/liveAdapters.ts) 做了按 `id` 去重和孤立边过滤。

如果后续又看到类似告警，优先检查：

- 后端是否再次返回重复 node/edge id
- adapter 去重逻辑是否被回退
- 聊天区 optimistic message 与服务端消息是否重复拼接

### 9.4 流式聊天兼容

前端当前兼容的是后端真实事件流，而不只是文档示例。

如果未来流式协议再改，优先检查：

- [src/api/conversations.ts](/Users/tylor/Code/GoAIPj/src/api/conversations.ts)
- [src/pages/ResourceGraphPage.tsx](/Users/tylor/Code/GoAIPj/src/pages/ResourceGraphPage.tsx)

### 9.5 communication 的维护规则

前后端通过：

- [communication.md](/Users/tylor/Code/GoAIPj/Communication/communication.md)

进行协作。

规则：

- 序号必须严格递增
- 超过 `10` 条后，必须先压缩成：
  - `front_end_privious_summary`
  - `back_end_privious_summary`
- 然后重新从 `communication time count 1` 开始

之前已经发生过“序号重复”和“说修了但联调进程不是最新代码”的问题，所以每次都要用真实请求复验，不要只看文字说明。

## 10. 当前任务清单状态

前端任务清单：

- [frontend_tasklist.md](/Users/tylor/Code/GoAIPj/docs/frontend_tasklist.md)

当前状态：

- `MVP` 已完成
- `Post-MVP` 已完成

已完成的 Post-MVP 包括：

- 深色模式
- 图谱导出 SVG
- 更丰富的图谱筛选器
- 会话持久化恢复

## 11. 建议的接手顺序

下一位开发如果继续接手，建议按这个顺序：

1. 先确认后端 `readyz`
2. 运行 `node src/test/liveSmoke.mjs`
3. 启动前端 `npm run dev`
4. 浏览关键页面：
   - `/groups`
   - `/groups/:groupId`
   - `/groups/:groupId/framework`
   - `/groups/:groupId/resources/:resourceId/graph`
5. 打开浏览器控制台，确认没有新的 runtime error / duplicate key warning
6. 再决定是否继续做新增需求，而不是先改现有 adapter

## 12. 如果接下来要继续开发

优先原则：

- 不要默认开 mock 掩盖真实问题
- 先复现，再在最靠近数据源的一层修问题
- 字段兼容优先放在 adapter 层
- 页面层尽量只做渲染和交互
- 每次涉及联调问题，都要把结论写进 [communication.md](/Users/tylor/Code/GoAIPj/Communication/communication.md)

如果只是验收或回归，最小动作集：

```bash
curl -sf http://127.0.0.1:8080/readyz
node src/test/liveSmoke.mjs
npm test
npm run build
```

## 13. 交接时的已知事实

- 当前前端默认走真实后端
- 当前真实后端地址按最近联调约定是 `127.0.0.1:8080`
- 当前后端 dev 进程是内存态
- 当前前端功能已可验收
- 当前没有已知新的前端阻塞问题
- 下一位接手时，首先应确认的不是页面代码，而是联调进程是否真的是最新后端进程
