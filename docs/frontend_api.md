# 学习知识图谱助手前端接口文档

## 1. 文档目标

本文档面向前端工程师，定义 V1 版本前端可调用的全部后端接口、调用方式、请求与响应结构、错误处理规范、SSE/流式调用方式，以及页面级调用建议。

本文档基于以下文档整理：

- [requirement.md](/Users/tylor/Code/GoAIPj/requirement.md)
- [front_requirement.md](/Users/tylor/Code/GoAIPj/front_requirement.md)
- [backend_design.md](/Users/tylor/Code/GoAIPj/backend_design.md)

## 2. 基础约定

### 2.1 Base URL

- 开发环境：`http://localhost:8080`
- API 前缀：`/api/v1`
- 完整基地址：`http://localhost:8080/api/v1`

### 2.2 协议与格式

- 普通接口协议：`HTTP/JSON`
- 文件上传：`multipart/form-data`
- 分组事件订阅：`text/event-stream`
- 流式问答：`text/event-stream`

### 2.3 当前版本认证约定

V1 默认单用户场景，接口暂不要求登录态。前端暂不需要传 `Authorization`。

后续如果接入认证，可统一通过请求头扩展：

```http
Authorization: Bearer <token>
```

### 2.4 通用请求头

普通 JSON 接口：

```http
Content-Type: application/json
Accept: application/json
```

文件上传接口：

- 不要手动设置 `Content-Type`
- 由浏览器自动生成 `multipart/form-data; boundary=...`

### 2.5 通用响应格式

成功：

```json
{
  "data": {},
  "error": null,
  "meta": {}
}
```

失败：

```json
{
  "data": null,
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "resource not found"
  },
  "meta": {}
}
```

### 2.6 前端统一错误处理建议

前端建议统一按以下规则处理：

1. `2xx` 视为成功。
2. `4xx` 视为用户输入或业务状态错误，直接提示。
3. `5xx` 视为服务异常，提示“服务繁忙，请稍后重试”。
4. 如果响应中 `error != null`，优先展示 `error.message`。

## 3. 通用 TypeScript 类型建议

```ts
export type ApiError = {
  code: string;
  message: string;
};

export type ApiResponse<T> = {
  data: T | null;
  error: ApiError | null;
  meta: Record<string, unknown>;
};

export type ID = string;
```

## 4. 枚举定义

### 4.1 资源类型

```ts
export type ResourceType =
  | "pdf"
  | "docx"
  | "pptx"
  | "xlsx"
  | "txt"
  | "md"
  | "web";
```

### 4.2 资源状态

```ts
export type ResourceStatus =
  | "uploaded"
  | "parsing"
  | "normalizing"
  | "graph_generating"
  | "completed"
  | "failed";
```

### 4.3 图谱类型

```ts
export type GraphType = "resource" | "framework";
```

### 4.4 节点类型

```ts
export type NodeType =
  | "topic"
  | "subtopic"
  | "concept"
  | "method"
  | "rule"
  | "conclusion"
  | "example";
```

### 4.5 节点来源类型

```ts
export type NodeSourceType = "extracted" | "summarized" | "expanded";
```

## 5. 业务对象结构

### 5.1 Group

```ts
export type Group = {
  id: ID;
  name: string;
  description: string | null;
  resource_count: number;
  created_at: string;
  updated_at?: string;
};
```

### 5.2 ResourceSummary

```ts
export type ResourceSummary = {
  id: ID;
  group_id: ID;
  name: string;
  resource_type: ResourceType;
  status: ResourceStatus;
  error_message: string | null;
  created_at: string;
  updated_at: string;
};
```

### 5.3 ResourceDetail

```ts
export type ResourceDetail = {
  id: ID;
  group_id: ID;
  name: string;
  resource_type: ResourceType;
  source_uri: string | null;
  status: ResourceStatus;
  error_code: string | null;
  error_message: string | null;
  failed_stage: string | null;
  latest_job_id: ID | null;
  latest_graph_id: ID | null;
  created_at: string;
  updated_at: string;
};
```

### 5.4 GraphNode

```ts
export type GraphNode = {
  id: ID;
  graph_id: ID;
  name: string;
  description: string | null;
  meaning: string | null;
  level: number;
  node_type: NodeType;
  source_type: NodeSourceType;
  is_expansion: boolean;
};
```

### 5.5 GraphEdge

```ts
export type GraphEdge = {
  id: ID;
  graph_id: ID;
  source_node_id: ID;
  target_node_id: ID;
  relation_type: string;
  relation_description: string | null;
  is_expansion_relation: boolean;
};
```

### 5.6 GraphView

```ts
export type GraphView = {
  graph: {
    id: ID;
    group_id: ID;
    resource_id: ID | null;
    graph_type: GraphType;
    status: "active" | "draft" | "archived" | "failed";
    version: number;
    summary: string | null;
    root_node_id: ID | null;
    created_at: string;
    updated_at: string;
  };
  nodes: GraphNode[];
  edges: GraphEdge[];
};
```

### 5.7 NodeDetail

```ts
export type NodeDetail = {
  id: ID;
  graph_id: ID;
  name: string;
  description: string | null;
  meaning: string | null;
  level: number;
  node_type: NodeType;
  source_type: NodeSourceType;
  is_expansion: boolean;
  neighbors: Array<{
    node_id: ID;
    node_name: string;
    relation_type: string;
    relation_description: string | null;
    direction: "in" | "out";
  }>;
  examples: Array<{
    id: ID;
    example_text: string;
    source_type: "extracted" | "generated";
  }>;
};
```

### 5.8 Conversation

```ts
export type Conversation = {
  id: ID;
  group_id: ID;
  graph_id: ID;
  current_node_id: ID;
  title: string | null;
  created_at: string;
  updated_at: string;
};
```

### 5.9 ConversationMessage

```ts
export type ConversationMessage = {
  id: ID;
  conversation_id: ID;
  current_node_id: ID;
  role: "user" | "assistant" | "system";
  content: string;
  citations: {
    chunk_ids: ID[];
    node_ids: ID[];
  };
  created_at: string;
};
```

### 5.10 Job

```ts
export type Job = {
  id: ID;
  job_type: string;
  status: "pending" | "running" | "succeeded" | "failed" | "retrying" | "cancelled";
  attempt: number;
  max_attempts: number;
  last_error: string | null;
  started_at: string | null;
  finished_at: string | null;
  created_at: string;
};
```

## 6. 前端接口总览

| 模块 | 方法 | 路径 | 用途 |
| --- | --- | --- | --- |
| 分组 | `POST` | `/groups` | 创建分组 |
| 分组 | `GET` | `/groups` | 获取分组列表 |
| 分组 | `GET` | `/groups/{groupId}` | 获取分组详情 |
| 分组 | `PATCH` | `/groups/{groupId}` | 更新分组 |
| 分组 | `DELETE` | `/groups/{groupId}` | 删除分组 |
| 资源 | `POST` | `/groups/{groupId}/resources` | 上传文件资源 |
| 资源 | `POST` | `/groups/{groupId}/web-resources` | 创建网页资源 |
| 资源 | `GET` | `/groups/{groupId}/resources` | 获取资源列表 |
| 资源 | `GET` | `/resources/{resourceId}` | 获取资源详情 |
| 资源 | `POST` | `/resources/{resourceId}/retry` | 重新处理资源 |
| 资源 | `DELETE` | `/resources/{resourceId}` | 删除资源 |
| 图谱 | `GET` | `/resources/{resourceId}/graph` | 获取资源级图谱 |
| 图谱 | `GET` | `/groups/{groupId}/framework-graph` | 获取框架图谱 |
| 图谱 | `POST` | `/groups/{groupId}/framework-graph/generate` | 手动生成框架图谱 |
| 节点 | `GET` | `/graphs/{graphId}/nodes/{nodeId}` | 获取节点详情 |
| 扩展 | `POST` | `/graphs/{graphId}/nodes/{nodeId}/expand` | 生成扩展节点 |
| 对话 | `POST` | `/conversations` | 创建会话 |
| 对话 | `GET` | `/conversations/{conversationId}` | 获取会话详情 |
| 对话 | `POST` | `/conversations/{conversationId}/messages` | 发送消息 |
| 任务 | `GET` | `/jobs/{jobId}` | 查询任务状态 |
| 事件 | `GET` | `/events/groups/{groupId}` | 订阅分组事件 |

## 7. 分组接口

### 7.1 创建分组

`POST /api/v1/groups`

请求体：

```json
{
  "name": "Go AI Project",
  "description": "项目学习资料"
}
```

响应：

```json
{
  "data": {
    "id": "grp_001",
    "name": "Go AI Project",
    "description": "项目学习资料",
    "resource_count": 0,
    "created_at": "2026-03-06T12:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 分组列表页点击“新建分组”

### 7.2 获取分组列表

`GET /api/v1/groups?page=1&page_size=20&keyword=go`

响应：

```json
{
  "data": {
    "items": [
      {
        "id": "grp_001",
        "name": "Go AI Project",
        "description": "项目学习资料",
        "resource_count": 3,
        "created_at": "2026-03-06T12:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 分组列表页初始化
- 搜索分组

### 7.3 获取分组详情

`GET /api/v1/groups/{groupId}`

响应：

```json
{
  "data": {
    "id": "grp_001",
    "name": "Go AI Project",
    "description": "项目学习资料",
    "resource_count": 3,
    "completed_resource_count": 2,
    "created_at": "2026-03-06T12:00:00Z",
    "updated_at": "2026-03-06T13:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 进入分组详情页时加载头部信息

### 7.4 更新分组

`PATCH /api/v1/groups/{groupId}`

请求体：

```json
{
  "name": "Go AI 学习",
  "description": "重命名后的描述"
}
```

### 7.5 删除分组

`DELETE /api/v1/groups/{groupId}`

响应：

```json
{
  "data": {
    "success": true
  },
  "error": null,
  "meta": {}
}
```

## 8. 资源接口

### 8.1 上传文件资源

`POST /api/v1/groups/{groupId}/resources`

请求方式：

- `multipart/form-data`

表单字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | `File` | 是 | 上传文件 |
| `name` | `string` | 否 | 自定义资源名，不传则使用文件名 |

成功响应：

```json
{
  "data": {
    "resource_id": "res_001",
    "status": "uploaded",
    "job_id": "job_001"
  },
  "error": null,
  "meta": {}
}
```

前端示例：

```ts
export async function uploadResource(groupId: string, file: File, name?: string) {
  const formData = new FormData();
  formData.append("file", file);
  if (name) formData.append("name", name);

  const res = await fetch(`/api/v1/groups/${groupId}/resources`, {
    method: "POST",
    body: formData,
  });

  return (await res.json()) as ApiResponse<{
    resource_id: string;
    status: ResourceStatus;
    job_id: string;
  }>;
}
```

前端调用场景：

- 分组详情页上传按钮

### 8.2 创建网页资源

`POST /api/v1/groups/{groupId}/web-resources`

请求体：

```json
{
  "url": "https://example.com/article",
  "name": "Go 并发文章"
}
```

响应：

```json
{
  "data": {
    "resource_id": "res_002",
    "status": "uploaded",
    "job_id": "job_002"
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 分组详情页“添加网页”

### 8.3 获取资源列表

`GET /api/v1/groups/{groupId}/resources`

支持查询参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `status` | `string` | 否 | 按状态过滤 |
| `page` | `number` | 否 | 分页页码 |
| `page_size` | `number` | 否 | 每页数量 |

响应：

```json
{
  "data": {
    "items": [
      {
        "id": "res_001",
        "group_id": "grp_001",
        "name": "system-design.pdf",
        "resource_type": "pdf",
        "status": "graph_generating",
        "error_message": null,
        "created_at": "2026-03-06T12:10:00Z",
        "updated_at": "2026-03-06T12:11:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 分组详情页资源表格

### 8.4 获取资源详情

`GET /api/v1/resources/{resourceId}`

响应：

```json
{
  "data": {
    "id": "res_001",
    "group_id": "grp_001",
    "name": "system-design.pdf",
    "resource_type": "pdf",
    "source_uri": null,
    "status": "completed",
    "error_code": null,
    "error_message": null,
    "failed_stage": null,
    "latest_job_id": "job_001",
    "latest_graph_id": "graph_001",
    "created_at": "2026-03-06T12:10:00Z",
    "updated_at": "2026-03-06T12:20:00Z"
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 资源图谱页顶部资源信息
- 失败详情展示

### 8.5 重新处理资源

`POST /api/v1/resources/{resourceId}/retry`

请求体：空

响应：

```json
{
  "data": {
    "resource_id": "res_001",
    "status": "uploaded",
    "job_id": "job_003"
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 资源失败后点击“重试”

### 8.6 删除资源

`DELETE /api/v1/resources/{resourceId}`

响应：

```json
{
  "data": {
    "success": true
  },
  "error": null,
  "meta": {}
}
```

## 9. 图谱接口

### 9.1 获取资源级知识图谱

`GET /api/v1/resources/{resourceId}/graph?include_expansion=true&max_level=3`

查询参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `include_expansion` | `boolean` | 否 | 是否返回扩展节点 |
| `max_level` | `number` | 否 | 最大层级，不传表示全部 |

响应：

```json
{
  "data": {
    "graph": {
      "id": "graph_001",
      "group_id": "grp_001",
      "resource_id": "res_001",
      "graph_type": "resource",
      "status": "active",
      "version": 1,
      "summary": "该文档主要介绍系统设计核心模块。",
      "root_node_id": "node_root",
      "created_at": "2026-03-06T12:20:00Z",
      "updated_at": "2026-03-06T12:20:00Z"
    },
    "nodes": [
      {
        "id": "node_root",
        "graph_id": "graph_001",
        "name": "系统设计",
        "description": "文档的总主题",
        "meaning": "用于组织后续子主题",
        "level": 1,
        "node_type": "topic",
        "source_type": "summarized",
        "is_expansion": false
      }
    ],
    "edges": []
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 资源图谱页初始化
- 切换层级过滤
- 切换“显示扩展节点”

### 9.2 获取分组框架图谱

`GET /api/v1/groups/{groupId}/framework-graph?max_level=2`

说明：

- 返回结构与资源图谱一致，只是 `graph.graph_type = "framework"`

前端调用场景：

- 分组详情页框架图谱区域
- 框架图谱独立页面

### 9.3 手动触发框架图谱生成

`POST /api/v1/groups/{groupId}/framework-graph/generate`

请求体：空

响应：

```json
{
  "data": {
    "group_id": "grp_001",
    "job_id": "job_100",
    "status": "pending"
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 分组详情页点击“重新生成框架图谱”

## 10. 节点详情与扩展接口

### 10.1 获取节点详情

`GET /api/v1/graphs/{graphId}/nodes/{nodeId}`

响应：

```json
{
  "data": {
    "id": "node_001",
    "graph_id": "graph_001",
    "name": "依赖注入",
    "description": "一种管理依赖关系的设计方式",
    "meaning": "在当前系统设计中用于降低模块耦合",
    "level": 2,
    "node_type": "concept",
    "source_type": "summarized",
    "is_expansion": false,
    "neighbors": [
      {
        "node_id": "node_002",
        "node_name": "模块解耦",
        "relation_type": "related",
        "relation_description": "依赖注入服务于模块解耦",
        "direction": "out"
      }
    ],
    "examples": [
      {
        "id": "ex_001",
        "example_text": "通过接口注入存储实现而非直接依赖具体数据库客户端。",
        "source_type": "generated"
      }
    ]
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 图谱节点点击后加载右侧详情面板

### 10.2 生成扩展节点

`POST /api/v1/graphs/{graphId}/nodes/{nodeId}/expand`

请求体：

```json
{
  "reason": "我想更理解这个概念的直接支撑知识"
}
```

响应：

```json
{
  "data": {
    "job_id": "job_expand_001",
    "status": "pending"
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 节点详情面板点击“扩展相关知识”

扩展完成后的前端处理建议：

1. 优先监听分组 SSE 事件。
2. 收到 `graph.node.expanded` 后重新拉取当前图谱。
3. 如果没有 SSE，可轮询 `GET /jobs/{jobId}`。

## 11. 对话接口

### 11.1 创建会话

`POST /api/v1/conversations`

请求体：

```json
{
  "group_id": "grp_001",
  "graph_id": "graph_001",
  "current_node_id": "node_001",
  "title": "依赖注入学习"
}
```

响应：

```json
{
  "data": {
    "id": "conv_001",
    "group_id": "grp_001",
    "graph_id": "graph_001",
    "current_node_id": "node_001",
    "title": "依赖注入学习",
    "created_at": "2026-03-06T12:30:00Z",
    "updated_at": "2026-03-06T12:30:00Z"
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 用户第一次打开某节点聊天面板时

### 11.2 获取会话详情

`GET /api/v1/conversations/{conversationId}`

响应：

```json
{
  "data": {
    "conversation": {
      "id": "conv_001",
      "group_id": "grp_001",
      "graph_id": "graph_001",
      "current_node_id": "node_001",
      "title": "依赖注入学习",
      "created_at": "2026-03-06T12:30:00Z",
      "updated_at": "2026-03-06T12:31:00Z"
    },
    "messages": [
      {
        "id": "msg_001",
        "conversation_id": "conv_001",
        "current_node_id": "node_001",
        "role": "user",
        "content": "这个节点在当前项目里有什么作用？",
        "citations": {
          "chunk_ids": [],
          "node_ids": []
        },
        "created_at": "2026-03-06T12:30:10Z"
      }
    ]
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 重新进入聊天面板恢复历史消息

### 11.3 发送消息，非流式

`POST /api/v1/conversations/{conversationId}/messages`

请求体：

```json
{
  "content": "这个节点在当前项目中的作用是什么？",
  "stream": false
}
```

响应：

```json
{
  "data": {
    "user_message": {
      "id": "msg_u_001",
      "conversation_id": "conv_001",
      "current_node_id": "node_001",
      "role": "user",
      "content": "这个节点在当前项目中的作用是什么？",
      "citations": {
        "chunk_ids": [],
        "node_ids": []
      },
      "created_at": "2026-03-06T12:31:00Z"
    },
    "assistant_message": {
      "id": "msg_a_001",
      "conversation_id": "conv_001",
      "current_node_id": "node_001",
      "role": "assistant",
      "content": "它的作用是帮助系统在模块之间解耦...",
      "citations": {
        "chunk_ids": ["chunk_001"],
        "node_ids": ["node_002"]
      },
      "created_at": "2026-03-06T12:31:02Z"
    }
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 不要求打字机效果时

### 11.4 发送消息，流式

`POST /api/v1/conversations/{conversationId}/messages`

请求体：

```json
{
  "content": "请用更容易理解的话解释这个节点",
  "stream": true
}
```

响应头：

```http
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

事件格式：

```txt
event: message.start
data: {"message_id":"msg_a_002"}

event: message.delta
data: {"delta":"它可以理解为把依赖从模块内部拿出来..."}

event: message.delta
data: {"delta":"这样每个模块只依赖接口而不是具体实现。"}

event: message.done
data: {
  "message": {
    "id": "msg_a_002",
    "conversation_id": "conv_001",
    "current_node_id": "node_001",
    "role": "assistant",
    "content": "它可以理解为把依赖从模块内部拿出来，这样每个模块只依赖接口而不是具体实现。",
    "citations": {
      "chunk_ids": ["chunk_001"],
      "node_ids": ["node_002"]
    },
    "created_at": "2026-03-06T12:32:00Z"
  }
}
```

前端流式调用示例：

```ts
export async function streamConversationMessage(
  conversationId: string,
  content: string,
  handlers: {
    onStart?: (messageId: string) => void;
    onDelta?: (delta: string) => void;
    onDone?: (message: ConversationMessage) => void;
    onError?: (error: unknown) => void;
  }
) {
  const response = await fetch(`/api/v1/conversations/${conversationId}/messages`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "text/event-stream",
    },
    body: JSON.stringify({ content, stream: true }),
  });

  const reader = response.body?.getReader();
  const decoder = new TextDecoder("utf-8");

  if (!reader) throw new Error("stream reader unavailable");

  let buffer = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const chunks = buffer.split("\n\n");
    buffer = chunks.pop() || "";

    for (const chunk of chunks) {
      const lines = chunk.split("\n");
      const event = lines.find((line) => line.startsWith("event:"))?.replace("event:", "").trim();
      const dataLine = lines.find((line) => line.startsWith("data:"))?.replace("data:", "").trim();
      if (!event || !dataLine) continue;

      const payload = JSON.parse(dataLine);

      if (event === "message.start") handlers.onStart?.(payload.message_id);
      if (event === "message.delta") handlers.onDelta?.(payload.delta);
      if (event === "message.done") handlers.onDone?.(payload.message);
    }
  }
}
```

## 12. 任务接口

### 12.1 查询任务状态

`GET /api/v1/jobs/{jobId}`

响应：

```json
{
  "data": {
    "id": "job_001",
    "job_type": "generate_resource_graph",
    "status": "running",
    "attempt": 1,
    "max_attempts": 3,
    "last_error": null,
    "started_at": "2026-03-06T12:11:00Z",
    "finished_at": null,
    "created_at": "2026-03-06T12:10:00Z"
  },
  "error": null,
  "meta": {}
}
```

前端调用场景：

- 没有 SSE 时作为兜底轮询
- 上传后查看单个任务状态

## 13. SSE 事件接口

### 13.1 订阅分组事件

`GET /api/v1/events/groups/{groupId}`

响应头：

```http
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

前端示例：

```ts
export function subscribeGroupEvents(groupId: string, onEvent: (event: string, payload: any) => void) {
  const source = new EventSource(`/api/v1/events/groups/${groupId}`);

  source.addEventListener("resource.status.changed", (e) => {
    onEvent("resource.status.changed", JSON.parse((e as MessageEvent).data));
  });

  source.addEventListener("resource.completed", (e) => {
    onEvent("resource.completed", JSON.parse((e as MessageEvent).data));
  });

  source.addEventListener("resource.failed", (e) => {
    onEvent("resource.failed", JSON.parse((e as MessageEvent).data));
  });

  source.addEventListener("framework_graph.updated", (e) => {
    onEvent("framework_graph.updated", JSON.parse((e as MessageEvent).data));
  });

  source.addEventListener("graph.node.expanded", (e) => {
    onEvent("graph.node.expanded", JSON.parse((e as MessageEvent).data));
  });

  return () => source.close();
}
```

### 13.2 事件类型与 payload

#### `resource.status.changed`

```json
{
  "group_id": "grp_001",
  "resource_id": "res_001",
  "status": "normalizing",
  "updated_at": "2026-03-06T12:12:00Z"
}
```

#### `resource.completed`

```json
{
  "group_id": "grp_001",
  "resource_id": "res_001",
  "graph_id": "graph_001",
  "status": "completed"
}
```

#### `resource.failed`

```json
{
  "group_id": "grp_001",
  "resource_id": "res_001",
  "status": "failed",
  "error_message": "pdf parse failed"
}
```

#### `framework_graph.updated`

```json
{
  "group_id": "grp_001",
  "graph_id": "framework_001",
  "status": "active"
}
```

#### `graph.node.expanded`

```json
{
  "group_id": "grp_001",
  "graph_id": "graph_001",
  "source_node_id": "node_001",
  "new_node_id": "node_010"
}
```

## 14. 页面级调用方式

## 14.1 分组列表页

进入页面时：

1. 调用 `GET /groups`

创建分组时：

1. 调用 `POST /groups`
2. 成功后刷新分组列表或直接插入本地缓存

重命名时：

1. 调用 `PATCH /groups/{groupId}`
2. 成功后更新本地缓存

删除时：

1. 调用 `DELETE /groups/{groupId}`
2. 成功后从列表移除

## 14.2 分组详情页

进入页面时：

1. 调用 `GET /groups/{groupId}`
2. 调用 `GET /groups/{groupId}/resources`
3. 调用 `GET /groups/{groupId}/framework-graph`
4. 打开 `GET /events/groups/{groupId}` 的 SSE 连接

上传文件时：

1. 调用 `POST /groups/{groupId}/resources`
2. 成功后先把资源插入列表，状态设为 `uploaded`
3. 通过 SSE 更新资源状态

添加网页时：

1. 调用 `POST /groups/{groupId}/web-resources`
2. 后续和文件上传相同

## 14.3 资源图谱页

进入页面时：

1. 调用 `GET /resources/{resourceId}`
2. 调用 `GET /resources/{resourceId}/graph`

点击节点时：

1. 调用 `GET /graphs/{graphId}/nodes/{nodeId}`

改变层级过滤时：

1. 重新调用 `GET /resources/{resourceId}/graph?max_level=...`

触发扩展时：

1. 调用 `POST /graphs/{graphId}/nodes/{nodeId}/expand`
2. 等待 SSE `graph.node.expanded`
3. 重新拉取图谱

## 14.4 节点对话面板

首次打开时：

1. 调用 `POST /conversations`
2. 保存 `conversationId`

再次打开时：

1. 调用 `GET /conversations/{conversationId}`

发送问题时：

1. 若需要流式体验，调用 `POST /conversations/{conversationId}/messages` 且 `stream=true`
2. 若只要完整结果，调用 `stream=false`

## 15. 推荐前端封装方式

### 15.1 API 目录建议

```text
src
├── api
│   ├── client.ts
│   ├── groups.ts
│   ├── resources.ts
│   ├── graphs.ts
│   ├── conversations.ts
│   ├── jobs.ts
│   └── events.ts
```

### 15.2 通用请求封装建议

```ts
export async function apiFetch<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const response = await fetch(input, init);
  const json = (await response.json()) as ApiResponse<T>;

  if (!response.ok || json.error) {
    throw new Error(json.error?.message || "Request failed");
  }

  return json.data as T;
}
```

### 15.3 TanStack Query key 建议

```ts
export const queryKeys = {
  groups: ["groups"] as const,
  groupDetail: (groupId: string) => ["group", groupId] as const,
  groupResources: (groupId: string) => ["group-resources", groupId] as const,
  frameworkGraph: (groupId: string, maxLevel?: number) => ["framework-graph", groupId, maxLevel] as const,
  resourceDetail: (resourceId: string) => ["resource", resourceId] as const,
  resourceGraph: (resourceId: string, params?: { maxLevel?: number; includeExpansion?: boolean }) =>
    ["resource-graph", resourceId, params] as const,
  nodeDetail: (graphId: string, nodeId: string) => ["node-detail", graphId, nodeId] as const,
  conversation: (conversationId: string) => ["conversation", conversationId] as const,
  job: (jobId: string) => ["job", jobId] as const,
};
```

## 16. 前端状态处理建议

### 16.1 资源状态文案映射

```ts
export const resourceStatusLabelMap: Record<ResourceStatus, string> = {
  uploaded: "已上传",
  parsing: "解析中",
  normalizing: "总结中",
  graph_generating: "图谱生成中",
  completed: "已完成",
  failed: "失败",
};
```

### 16.2 图谱页面空态建议

1. `resource.status !== completed` 时展示处理中状态，不直接报错。
2. `resource.status === failed` 时展示失败原因和重试按钮。
3. `graph` 为空但资源未完成时，继续等待 SSE 或轮询。

## 17. 常见错误码建议

| 错误码 | 含义 | 前端处理建议 |
| --- | --- | --- |
| `GROUP_NOT_FOUND` | 分组不存在 | 返回分组列表页 |
| `RESOURCE_NOT_FOUND` | 资源不存在 | 提示并返回分组详情页 |
| `GRAPH_NOT_FOUND` | 图谱不存在 | 显示“图谱尚未生成” |
| `NODE_NOT_FOUND` | 节点不存在 | 清空节点详情面板 |
| `INVALID_RESOURCE_TYPE` | 文件类型不支持 | 上传前后都可提示 |
| `RESOURCE_PROCESSING_CONFLICT` | 当前资源已有运行中任务 | 提示稍后查看 |
| `FRAMEWORK_GRAPH_GENERATING` | 框架图谱正在生成 | 按处理中提示 |
| `CONVERSATION_NOT_FOUND` | 会话不存在 | 重新创建会话 |
| `INVALID_URL` | 网页地址不合法 | 表单提示 |

## 18. 实施建议

1. 前端优先接入 `分组 + 资源 + 图谱 + 节点详情 + 对话 + SSE` 六类接口。
2. 资源状态更新优先用 SSE，不建议仅靠高频轮询。
3. 图谱查询接口建议始终通过 Query 参数控制层级，不要在前端自己裁节点后再推导边。
4. 流式对话建议单独做 API 封装，不要复用普通 JSON 请求封装。

## 19. 结论

这份接口文档已经覆盖了 V1 前端工作台所需的全部核心接口，以及调用方式、请求参数、响应结构、页面接入方式和 SSE/流式问答接入规范。前端可以直接基于本文档建立 `api client`、`TanStack Query hooks` 和页面级数据流，而不需要再从后端设计文档中二次推导接口细节。
