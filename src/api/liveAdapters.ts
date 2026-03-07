import type {
  Conversation,
  ConversationDetail,
  ConversationMessage,
  GraphEdge,
  GraphNode,
  GraphView,
  Group,
  GroupDetail,
  NodeDetail,
  Paginated,
  ResourceDetail,
  ResourceSummary,
  ResourceType,
  SendMessageResponse,
} from "@/api/types";

type MaybePaginated<T> = Paginated<T> | T[];

function dedupeGraphNodes(items: GraphNode[]) {
  return Array.from(new Map(items.map((item) => [item.id, item])).values());
}

function dedupeGraphEdges(items: GraphEdge[]) {
  return Array.from(new Map(items.map((item) => [item.id, item])).values());
}

function asResourceType(value: unknown): ResourceType {
  const normalized = typeof value === "string" ? value.toLowerCase() : "md";
  return normalized === "html" ? "web" : (normalized as ResourceType);
}

export function normalizePaginated<T>(value: MaybePaginated<T>): Paginated<T> {
  if (Array.isArray(value)) {
    return {
      items: value,
      total: value.length,
      page: 1,
      page_size: value.length || 20,
    };
  }

  return value;
}

export function normalizeGroup(value: any): Group {
  return {
    id: value.id,
    name: value.name,
    description: value.description ?? null,
    resource_count: value.resource_count ?? 0,
    created_at: value.created_at,
    updated_at: value.updated_at ?? undefined,
  };
}

export function normalizeGroupDetail(value: any): GroupDetail {
  return {
    ...normalizeGroup(value),
    completed_resource_count: value.completed_resource_count ?? 0,
  };
}

export function normalizeGroups(value: MaybePaginated<any>): Paginated<Group> {
  const paginated = normalizePaginated(value);
  return {
    ...paginated,
    items: paginated.items.map(normalizeGroup),
  };
}

export function normalizeResourceSummary(value: any): ResourceSummary {
  return {
    id: value.id,
    group_id: value.group_id,
    name: value.name,
    resource_type: asResourceType(value.resource_type ?? value.type),
    status: value.status,
    error_message: value.error_message ?? null,
    created_at: value.created_at,
    updated_at: value.updated_at,
  };
}

export function normalizeResourceSummaries(value: MaybePaginated<any>): Paginated<ResourceSummary> {
  const paginated = normalizePaginated(value);
  return {
    ...paginated,
    items: paginated.items.map(normalizeResourceSummary),
  };
}

export function normalizeResourceDetail(value: any): ResourceDetail {
  return {
    id: value.id,
    group_id: value.group_id,
    name: value.name,
    resource_type: asResourceType(value.resource_type ?? value.type),
    source_uri: value.source_uri ?? value.source_url ?? null,
    status: value.status,
    error_code: value.error_code ?? null,
    error_message: value.error_message ?? null,
    failed_stage: value.failed_stage ?? null,
    latest_job_id: value.latest_job_id ?? null,
    latest_graph_id: value.latest_graph_id ?? null,
    created_at: value.created_at,
    updated_at: value.updated_at,
  };
}

function inferRootNodeId(nodes: any[]) {
  return nodes.find((node) => node.id === "root")?.id ?? nodes.find((node) => node.level === 0)?.id ?? nodes[0]?.id ?? null;
}

export function normalizeGraphView(value: any): GraphView {
  const nodes = dedupeGraphNodes((value.nodes ?? []).map((node: any) => ({
    id: node.id,
    graph_id: node.graph_id,
    name: node.name,
    description: node.description ?? null,
    meaning: node.meaning ?? null,
    level: node.level === 0 ? 1 : node.level,
    node_type: node.node_type ?? node.type ?? "topic",
    source_type: node.source_type ?? "summarized",
    is_expansion: node.is_expansion ?? false,
  })));

  const nodeIds = new Set(nodes.map((node) => node.id));
  const edges = dedupeGraphEdges(
    (value.edges ?? []).map((edge: any) => ({
      id: edge.id,
      graph_id: edge.graph_id,
      source_node_id: edge.source_node_id ?? edge.source_id,
      target_node_id: edge.target_node_id ?? edge.target_id,
      relation_type: edge.relation_type ?? edge.relation ?? "related",
      relation_description: edge.relation_description ?? null,
      is_expansion_relation: edge.is_expansion_relation ?? edge.is_expansion ?? false,
    })),
  ).filter((edge) => nodeIds.has(edge.source_node_id) && nodeIds.has(edge.target_node_id));

  return {
    graph: {
      id: value.graph.id,
      group_id: value.graph.group_id,
      resource_id: value.graph.resource_id || null,
      graph_type: value.graph.graph_type ?? (value.graph.resource_id ? "resource" : "framework"),
      status: value.graph.status ?? (value.graph.is_active === false ? "draft" : "active"),
      version: value.graph.version ?? 1,
      summary: value.graph.summary ?? value.graph.title ?? null,
      root_node_id: value.graph.root_node_id ?? inferRootNodeId(nodes),
      created_at: value.graph.created_at,
      updated_at: value.graph.updated_at,
    },
    nodes,
    edges,
  };
}

export function normalizeMessage(value: any, fallbackNodeId = ""): ConversationMessage {
  const contextNodeId = value.context_snapshot?.current_node_id;
  return {
    id: value.id,
    conversation_id: value.conversation_id,
    current_node_id: value.current_node_id ?? contextNodeId ?? fallbackNodeId,
    role: value.role,
    content: value.content,
    citations: {
      chunk_ids: value.citations?.chunk_ids ?? value.cited_chunk_ids ?? [],
      node_ids: value.citations?.node_ids ?? value.cited_node_ids ?? [],
    },
    created_at: value.created_at,
  };
}

export function normalizeConversation(value: any): Conversation {
  return {
    id: value.id,
    group_id: value.group_id,
    graph_id: value.graph_id,
    current_node_id: value.current_node_id,
    title: value.title ?? null,
    created_at: value.created_at,
    updated_at: value.updated_at,
  };
}

export function normalizeConversationDetail(value: any): ConversationDetail {
  const conversation = normalizeConversation(value.conversation ?? value);
  const messages = (value.messages ?? value.conversation?.messages ?? value.messages ?? value.data?.messages ?? []).map((message: any) =>
    normalizeMessage(message, conversation.current_node_id),
  );

  return { conversation, messages };
}

export function normalizeSendMessageResponse(value: any, fallbackNodeId = ""): SendMessageResponse {
  return {
    user_message: normalizeMessage(value.user_message, fallbackNodeId),
    assistant_message: normalizeMessage(value.assistant_message, fallbackNodeId),
  };
}

export function normalizeNodeDetail(value: any): NodeDetail {
  const node = value.node ?? value;
  return {
    id: node.id,
    graph_id: node.graph_id,
    name: node.name,
    description: node.description ?? null,
    meaning: node.meaning ?? null,
    level: node.level === 0 ? 1 : node.level,
    node_type: node.node_type ?? node.type ?? "topic",
    source_type: node.source_type ?? "summarized",
    is_expansion: node.is_expansion ?? false,
    neighbors: (value.neighbors ?? []).map((neighbor: any) => ({
      node_id: neighbor.node_id ?? neighbor.node?.id,
      node_name: neighbor.node_name ?? neighbor.node?.name,
      relation_type: neighbor.relation_type ?? neighbor.relation ?? "related",
      relation_description: neighbor.relation_description ?? null,
      direction: neighbor.direction ?? "out",
    })),
    examples: (value.examples ?? []).map((example: any) => ({
      id: example.id,
      example_text: example.example_text ?? example.content ?? "",
      source_type: example.source_type ?? "generated",
    })),
  };
}
