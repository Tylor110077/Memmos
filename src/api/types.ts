export type ID = string;

export type ApiError = {
  code: string;
  message: string;
};

export type ApiResponse<T> = {
  data: T | null;
  error: ApiError | null;
  meta: Record<string, unknown>;
};

export type ResourceType = "pdf" | "docx" | "pptx" | "xlsx" | "txt" | "md" | "web";
export type ResourceStatus =
  | "uploaded"
  | "parsing"
  | "normalizing"
  | "graph_generating"
  | "completed"
  | "failed";
export type GraphType = "resource" | "framework";
export type NodeType = "topic" | "subtopic" | "concept" | "method" | "rule" | "conclusion" | "example";
export type NodeSourceType = "extracted" | "summarized" | "expanded";

export type Group = {
  id: ID;
  name: string;
  description: string | null;
  resource_count: number;
  created_at: string;
  updated_at?: string;
};

export type GroupDetail = Group & {
  completed_resource_count: number;
};

export type Paginated<T> = {
  items: T[];
  total: number;
  page: number;
  page_size: number;
};

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

export type GraphEdge = {
  id: ID;
  graph_id: ID;
  source_node_id: ID;
  target_node_id: ID;
  relation_type: string;
  relation_description: string | null;
  is_expansion_relation: boolean;
};

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

export type Conversation = {
  id: ID;
  group_id: ID;
  graph_id: ID;
  current_node_id: ID;
  title: string | null;
  created_at: string;
  updated_at: string;
};

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

export type ConversationDetail = {
  conversation: Conversation;
  messages: ConversationMessage[];
};

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

export type ResourceMutationResult = {
  resource_id: ID;
  status: ResourceStatus;
  job_id: ID;
};

export type FrameworkGraphGenerateResult = {
  group_id: ID;
  job_id: ID;
  status: "pending";
};

export type ExpandNodeResult = {
  job_id: ID;
  status: "pending";
};

export type SendMessageResponse = {
  user_message: ConversationMessage;
  assistant_message: ConversationMessage;
};

export type GroupEventPayload =
  | {
      group_id: ID;
      resource_id: ID;
      status: ResourceStatus;
      updated_at: string;
    }
  | {
      group_id: ID;
      resource_id: ID;
      graph_id: ID;
      status: "completed";
    }
  | {
      group_id: ID;
      resource_id: ID;
      status: "failed";
      error_message: string;
    }
  | {
      group_id: ID;
      graph_id: ID;
      status: "active";
    }
  | {
      group_id: ID;
      graph_id: ID;
      source_node_id: ID;
      new_node_id: ID;
    };
