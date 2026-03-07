import type {
  Conversation,
  ConversationMessage,
  GraphView,
  GroupDetail,
  GroupEventPayload,
  Job,
  NodeDetail,
  ResourceDetail,
  ResourceSummary,
} from "@/api/types";

const now = "2026-03-06T12:00:00Z";

function currentTimestamp() {
  return new Date().toISOString();
}

export let groups: GroupDetail[] = [
  {
    id: "grp_agent",
    name: "LLM Agent 体系",
    description: "围绕 Agent 架构、规划、执行和工具调用能力建立统一学习框架。",
    resource_count: 4,
    completed_resource_count: 2,
    created_at: now,
    updated_at: "2026-03-06T13:00:00Z",
  },
  {
    id: "grp_mm",
    name: "多模态检索设计",
    description: "聚合 PDF、网页和 Markdown，整理为学习导图。",
    resource_count: 3,
    completed_resource_count: 1,
    created_at: "2026-03-05T10:00:00Z",
    updated_at: "2026-03-05T16:00:00Z",
  },
  {
    id: "grp_go",
    name: "Go + Eino 实践",
    description: "整理 API 编排、流程节点、消息上下文与执行链路。",
    resource_count: 5,
    completed_resource_count: 4,
    created_at: "2026-03-04T08:00:00Z",
    updated_at: "2026-03-06T11:00:00Z",
  },
];

export let resources: ResourceDetail[] = [
  {
    id: "res_cookbook",
    group_id: "grp_agent",
    name: "OpenAI Agents Cookbook.pdf",
    resource_type: "pdf",
    source_uri: null,
    status: "completed",
    error_code: null,
    error_message: null,
    failed_stage: null,
    latest_job_id: "job_cookbook",
    latest_graph_id: "graph_cookbook",
    created_at: "2026-03-06T10:12:00Z",
    updated_at: "2026-03-06T10:20:00Z",
  },
  {
    id: "res_memory",
    group_id: "grp_agent",
    name: "Agent Memory System.md",
    resource_type: "md",
    source_uri: null,
    status: "completed",
    error_code: null,
    error_message: null,
    failed_stage: null,
    latest_job_id: "job_memory",
    latest_graph_id: "graph_memory",
    created_at: "2026-03-06T09:40:00Z",
    updated_at: "2026-03-06T09:52:00Z",
  },
  {
    id: "res_planner",
    group_id: "grp_agent",
    name: "Planner vs Executor.html",
    resource_type: "web",
    source_uri: "https://example.com/planner-executor",
    status: "graph_generating",
    error_code: null,
    error_message: null,
    failed_stage: null,
    latest_job_id: "job_planner",
    latest_graph_id: "graph_planner",
    created_at: "2026-03-06T08:55:00Z",
    updated_at: "2026-03-06T09:00:00Z",
  },
  {
    id: "res_function",
    group_id: "grp_agent",
    name: "Function Calling Patterns.pptx",
    resource_type: "pptx",
    source_uri: null,
    status: "failed",
    error_code: "PPT_PARSE_FAILED",
    error_message: "ppt parse failed",
    failed_stage: "parsing",
    latest_job_id: "job_function",
    latest_graph_id: null,
    created_at: "2026-03-05T16:10:00Z",
    updated_at: "2026-03-05T16:18:00Z",
  },
];

export const graphViews: Record<string, GraphView> = {
  graph_cookbook: {
    graph: {
      id: "graph_cookbook",
      group_id: "grp_agent",
      resource_id: "res_cookbook",
      graph_type: "resource",
      status: "active",
      version: 1,
      summary: "该资源围绕 Agent 执行闭环、工具调用与记忆机制展开。",
      root_node_id: "node_loop",
      created_at: now,
      updated_at: now,
    },
    nodes: [
      {
        id: "node_loop",
        graph_id: "graph_cookbook",
        name: "Agent 执行循环",
        description: "代理接收目标、形成计划、调用工具并观察反馈的连续过程。",
        meaning: "连接规划、工具调用和记忆上下文，是理解 Agent 行为闭环的关键。",
        level: 1,
        node_type: "topic",
        source_type: "summarized",
        is_expansion: false,
      },
      {
        id: "node_plan",
        graph_id: "graph_cookbook",
        name: "任务规划",
        description: "将高层目标拆分为可执行步骤。",
        meaning: "帮助执行循环聚焦下一步行为。",
        level: 2,
        node_type: "subtopic",
        source_type: "summarized",
        is_expansion: false,
      },
      {
        id: "node_tool",
        graph_id: "graph_cookbook",
        name: "工具调用",
        description: "把内部推理转化为对外部能力的访问。",
        meaning: "连接模型与外部系统。",
        level: 2,
        node_type: "method",
        source_type: "summarized",
        is_expansion: false,
      },
      {
        id: "node_observe",
        graph_id: "graph_cookbook",
        name: "Observation 反馈",
        description: "读取工具结果和环境变化。",
        meaning: "让下一轮决策有依据。",
        level: 3,
        node_type: "concept",
        source_type: "extracted",
        is_expansion: false,
      },
      {
        id: "node_memory",
        graph_id: "graph_cookbook",
        name: "Memory 注入",
        description: "将历史上下文与当前任务拼接。",
        meaning: "保证多轮学习过程连续。",
        level: 3,
        node_type: "concept",
        source_type: "summarized",
        is_expansion: false,
      },
      {
        id: "node_retry",
        graph_id: "graph_cookbook",
        name: "扩展：重试策略",
        description: "工具调用失败后的重试与降级策略。",
        meaning: "帮助理解鲁棒性设计。",
        level: 3,
        node_type: "rule",
        source_type: "expanded",
        is_expansion: true,
      },
    ],
    edges: [
      {
        id: "edge_1",
        graph_id: "graph_cookbook",
        source_node_id: "node_plan",
        target_node_id: "node_loop",
        relation_type: "supports",
        relation_description: "规划支撑执行循环",
        is_expansion_relation: false,
      },
      {
        id: "edge_2",
        graph_id: "graph_cookbook",
        source_node_id: "node_loop",
        target_node_id: "node_tool",
        relation_type: "contains",
        relation_description: "执行循环包含工具调用阶段",
        is_expansion_relation: false,
      },
      {
        id: "edge_3",
        graph_id: "graph_cookbook",
        source_node_id: "node_loop",
        target_node_id: "node_observe",
        relation_type: "contains",
        relation_description: "执行循环包含观察反馈",
        is_expansion_relation: false,
      },
      {
        id: "edge_4",
        graph_id: "graph_cookbook",
        source_node_id: "node_loop",
        target_node_id: "node_memory",
        relation_type: "contains",
        relation_description: "执行循环依赖上下文记忆",
        is_expansion_relation: false,
      },
      {
        id: "edge_5",
        graph_id: "graph_cookbook",
        source_node_id: "node_tool",
        target_node_id: "node_retry",
        relation_type: "extends",
        relation_description: "工具调用扩展到重试策略",
        is_expansion_relation: true,
      },
    ],
  },
  framework_agent: {
    graph: {
      id: "framework_agent",
      group_id: "grp_agent",
      resource_id: null,
      graph_type: "framework",
      status: "active",
      version: 1,
      summary: "高层框架用于建立 Agent 体系整体认知。",
      root_node_id: "f_root",
      created_at: now,
      updated_at: now,
    },
    nodes: [
      {
        id: "f_root",
        graph_id: "framework_agent",
        name: "LLM Agent 体系",
        description: "分组级总览主题。",
        meaning: "汇总多个资源中共性的核心结构。",
        level: 1,
        node_type: "topic",
        source_type: "summarized",
        is_expansion: false,
      },
      {
        id: "f_plan",
        graph_id: "framework_agent",
        name: "规划与分解",
        description: "目标理解与步骤拆解。",
        meaning: null,
        level: 2,
        node_type: "subtopic",
        source_type: "summarized",
        is_expansion: false,
      },
      {
        id: "f_exec",
        graph_id: "framework_agent",
        name: "执行与反馈",
        description: "执行链路与观察反馈闭环。",
        meaning: null,
        level: 2,
        node_type: "subtopic",
        source_type: "summarized",
        is_expansion: false,
      },
      {
        id: "f_tool",
        graph_id: "framework_agent",
        name: "工具调用",
        description: "对外部能力的访问抽象。",
        meaning: null,
        level: 2,
        node_type: "method",
        source_type: "summarized",
        is_expansion: false,
      },
      {
        id: "f_memory",
        graph_id: "framework_agent",
        name: "记忆与上下文",
        description: "会话记忆与长期知识。",
        meaning: null,
        level: 2,
        node_type: "concept",
        source_type: "summarized",
        is_expansion: false,
      },
    ],
    edges: [
      {
        id: "f_edge_1",
        graph_id: "framework_agent",
        source_node_id: "f_root",
        target_node_id: "f_plan",
        relation_type: "contains",
        relation_description: null,
        is_expansion_relation: false,
      },
      {
        id: "f_edge_2",
        graph_id: "framework_agent",
        source_node_id: "f_root",
        target_node_id: "f_exec",
        relation_type: "contains",
        relation_description: null,
        is_expansion_relation: false,
      },
      {
        id: "f_edge_3",
        graph_id: "framework_agent",
        source_node_id: "f_root",
        target_node_id: "f_tool",
        relation_type: "contains",
        relation_description: null,
        is_expansion_relation: false,
      },
      {
        id: "f_edge_4",
        graph_id: "framework_agent",
        source_node_id: "f_root",
        target_node_id: "f_memory",
        relation_type: "contains",
        relation_description: null,
        is_expansion_relation: false,
      },
    ],
  },
};

export const nodeDetails: Record<string, NodeDetail> = {
  node_loop: {
    id: "node_loop",
    graph_id: "graph_cookbook",
    name: "Agent 执行循环",
    description: "表示代理从接收目标、形成计划、调用工具到观察反馈并决定下一步动作的连续过程。",
    meaning: "它是整个资源的主干节点，连接规划、工具调用和记忆上下文，是理解 Agent 行为闭环的关键。",
    level: 1,
    node_type: "topic",
    source_type: "summarized",
    is_expansion: false,
    neighbors: [
      {
        node_id: "node_tool",
        node_name: "工具调用",
        relation_type: "contains",
        relation_description: "执行循环包含工具调用阶段",
        direction: "out",
      },
      {
        node_id: "node_plan",
        node_name: "任务规划",
        relation_type: "supports",
        relation_description: "规划支撑执行循环",
        direction: "in",
      },
    ],
    examples: [
      {
        id: "ex_1",
        example_text: "先检索知识库，再根据返回内容调用函数，随后将结果写入会话记忆。",
        source_type: "generated",
      },
    ],
  },
  node_tool: {
    id: "node_tool",
    graph_id: "graph_cookbook",
    name: "工具调用",
    description: "把内部推理转化为外部系统能力调用。",
    meaning: "使 Agent 可以访问检索、执行和外部系统。",
    level: 2,
    node_type: "method",
    source_type: "summarized",
    is_expansion: false,
    neighbors: [
      {
        node_id: "node_loop",
        node_name: "Agent 执行循环",
        relation_type: "belongs_to",
        relation_description: "工具调用属于执行循环",
        direction: "in",
      },
      {
        node_id: "node_retry",
        node_name: "扩展：重试策略",
        relation_type: "extends",
        relation_description: "工具调用可扩展出重试策略",
        direction: "out",
      },
    ],
    examples: [
      {
        id: "ex_2",
        example_text: "根据规划结果调用函数获取外部数据，再将结果返回给模型。",
        source_type: "extracted",
      },
    ],
  },
};

export let conversations: Record<string, { conversation: Conversation; messages: ConversationMessage[] }> = {
  conv_loop: {
    conversation: {
      id: "conv_loop",
      group_id: "grp_agent",
      graph_id: "graph_cookbook",
      current_node_id: "node_loop",
      title: "Agent 执行循环学习",
      created_at: now,
      updated_at: now,
    },
    messages: [
      {
        id: "msg_sys",
        conversation_id: "conv_loop",
        current_node_id: "node_loop",
        role: "system",
        content: "当前回答会围绕“Agent 执行循环”，并结合本分组背景知识解释。",
        citations: { chunk_ids: [], node_ids: [] },
        created_at: now,
      },
      {
        id: "msg_user_1",
        conversation_id: "conv_loop",
        current_node_id: "node_loop",
        role: "user",
        content: "它和工具调用节点是什么关系？",
        citations: { chunk_ids: [], node_ids: [] },
        created_at: now,
      },
      {
        id: "msg_ai_1",
        conversation_id: "conv_loop",
        current_node_id: "node_loop",
        role: "assistant",
        content: "工具调用是执行循环中的关键阶段，用于把内部推理转化为外部能力访问。",
        citations: { chunk_ids: ["chunk_001"], node_ids: ["node_tool"] },
        created_at: now,
      },
    ],
  },
};

export let jobs: Job[] = [
  {
    id: "job_cookbook",
    job_type: "generate_resource_graph",
    status: "succeeded",
    attempt: 1,
    max_attempts: 3,
    last_error: null,
    started_at: now,
    finished_at: now,
    created_at: now,
  },
];

export const groupEvents: GroupEventPayload[] = [
  {
    group_id: "grp_agent",
    resource_id: "res_planner",
    status: "graph_generating",
    updated_at: now,
  },
];

export function listResources(groupId: string): ResourceSummary[] {
  return resources
    .filter((resource) => resource.group_id === groupId)
    .map(({ error_message, ...rest }) => ({
      ...rest,
      error_message,
    }));
}

export function addGroup(payload: { name: string; description?: string }) {
  const timestamp = currentTimestamp();
  const group: GroupDetail = {
    id: `grp_${Date.now()}`,
    name: payload.name,
    description: payload.description ?? null,
    resource_count: 0,
    completed_resource_count: 0,
    created_at: timestamp,
    updated_at: timestamp,
  };
  groups = [group, ...groups];
  return group;
}

export function patchGroup(groupId: string, payload: { name: string; description?: string }) {
  const timestamp = currentTimestamp();
  groups = groups.map((group) =>
    group.id === groupId
      ? { ...group, name: payload.name, description: payload.description ?? null, updated_at: timestamp }
      : group,
  );
  return groups.find((group) => group.id === groupId) ?? null;
}

export function removeGroup(groupId: string) {
  groups = groups.filter((group) => group.id !== groupId);
}

export function addWebResource(groupId: string, payload: { url: string; name?: string }) {
  const timestamp = currentTimestamp();
  const detail: ResourceDetail = {
    id: `res_${Date.now()}`,
    group_id: groupId,
    name: payload.name || payload.url,
    resource_type: "web",
    source_uri: payload.url,
    status: "uploaded",
    error_code: null,
    error_message: null,
    failed_stage: null,
    latest_job_id: `job_${Date.now()}`,
    latest_graph_id: null,
    created_at: timestamp,
    updated_at: timestamp,
  };
  resources = [detail, ...resources];
  groups = groups.map((group) => (group.id === groupId ? { ...group, resource_count: group.resource_count + 1 } : group));
  return detail;
}

export function addUploadResource(groupId: string, name: string) {
  const timestamp = currentTimestamp();
  const detail: ResourceDetail = {
    id: `res_${Date.now()}`,
    group_id: groupId,
    name,
    resource_type: "pdf",
    source_uri: null,
    status: "uploaded",
    error_code: null,
    error_message: null,
    failed_stage: null,
    latest_job_id: `job_${Date.now()}`,
    latest_graph_id: null,
    created_at: timestamp,
    updated_at: timestamp,
  };
  resources = [detail, ...resources];
  groups = groups.map((group) => (group.id === groupId ? { ...group, resource_count: group.resource_count + 1 } : group));
  return detail;
}

export function removeResource(resourceId: string) {
  const target = resources.find((resource) => resource.id === resourceId);
  resources = resources.filter((resource) => resource.id !== resourceId);
  if (target) {
    groups = groups.map((group) =>
      group.id === target.group_id ? { ...group, resource_count: Math.max(group.resource_count - 1, 0) } : group,
    );
  }
}

export function retryResourceJob(resourceId: string) {
  const timestamp = currentTimestamp();
  resources = resources.map((resource) =>
    resource.id === resourceId
      ? {
          ...resource,
          status: "uploaded",
          error_code: null,
          error_message: null,
          failed_stage: null,
          latest_job_id: `job_${Date.now()}`,
          updated_at: timestamp,
        }
      : resource,
  );
  return resources.find((resource) => resource.id === resourceId) ?? null;
}

export function appendConversationMessage(conversationId: string, currentNodeId: string, content: string) {
  const timestamp = currentTimestamp();
  const conversationEntry = conversations[conversationId];
  const userMessage: ConversationMessage = {
    id: `msg_user_${Date.now()}`,
    conversation_id: conversationId,
    current_node_id: currentNodeId,
    role: "user",
    content,
    citations: { chunk_ids: [], node_ids: [] },
    created_at: timestamp,
  };
  const assistantMessage: ConversationMessage = {
    id: `msg_ai_${Date.now()}`,
    conversation_id: conversationId,
    current_node_id: currentNodeId,
    role: "assistant",
    content: "这个节点在当前知识结构中承担主干解释作用，并把相关概念串成可追问的学习链路。",
    citations: { chunk_ids: ["chunk_001"], node_ids: ["node_tool"] },
    created_at: timestamp,
  };
  conversationEntry.messages = [...conversationEntry.messages, userMessage, assistantMessage];
  return { user_message: userMessage, assistant_message: assistantMessage };
}
