import { HttpResponse, http } from "msw";
import type { Conversation, ConversationDetail, SendMessageResponse } from "@/api/types";
import {
  addGroup,
  addUploadResource,
  addWebResource,
  appendConversationMessage,
  conversations,
  graphViews,
  groups,
  jobs,
  listResources,
  nodeDetails,
  patchGroup,
  removeGroup,
  removeResource,
  resources,
  retryResourceJob,
} from "@/mocks/data";

function ok<T>(data: T) {
  return HttpResponse.json({ data, error: null, meta: {} });
}

function fail(code: string, message: string, status = 400) {
  return HttpResponse.json(
    { data: null, error: { code, message }, meta: {} },
    { status },
  );
}

export const handlers = [
  http.get("/api/v1/groups", ({ request }) => {
    const url = new URL(request.url);
    const keyword = url.searchParams.get("keyword")?.trim() ?? "";
    const items = keyword ? groups.filter((group) => group.name.includes(keyword)) : groups;
    return ok({
      items,
      total: items.length,
      page: 1,
      page_size: 20,
    });
  }),

  http.post("/api/v1/groups", async ({ request }) => {
    const body = (await request.json()) as { name?: string; description?: string };
    if (!body.name?.trim()) {
      return fail("INVALID_ARGUMENT", "name is required");
    }
    if (body.name.length > 30) {
      return fail("INVALID_ARGUMENT", "name length should be <= 30");
    }
    return ok(addGroup({ name: body.name, description: body.description }));
  }),

  http.get("/api/v1/groups/:groupId", ({ params }) => {
    const group = groups.find((item) => item.id === params.groupId);
    return group ? ok(group) : fail("RESOURCE_NOT_FOUND", "group not found", 404);
  }),

  http.patch("/api/v1/groups/:groupId", async ({ params, request }) => {
    const body = (await request.json()) as { name?: string; description?: string };
    if (!body.name?.trim()) {
      return fail("INVALID_ARGUMENT", "name is required");
    }
    const group = patchGroup(String(params.groupId), { name: body.name, description: body.description });
    return group ? ok(group) : fail("RESOURCE_NOT_FOUND", "group not found", 404);
  }),

  http.delete("/api/v1/groups/:groupId", ({ params }) => {
    removeGroup(String(params.groupId));
    return ok({ success: true });
  }),

  http.get("/api/v1/groups/:groupId/resources", ({ params }) => {
    const items = listResources(String(params.groupId));
    return ok({
      items,
      total: items.length,
      page: 1,
      page_size: 20,
    });
  }),

  http.post("/api/v1/groups/:groupId/resources", async ({ params, request }) => {
    const formData = await request.formData();
    const file = formData.get("file");
    if (!(file instanceof File)) {
      return fail("INVALID_ARGUMENT", "file is required");
    }
    const detail = addUploadResource(String(params.groupId), String(formData.get("name") || file.name));
    return ok({
      resource_id: detail.id,
      status: detail.status,
      job_id: detail.latest_job_id!,
    });
  }),

  http.post("/api/v1/groups/:groupId/web-resources", async ({ params, request }) => {
    const body = (await request.json()) as { url?: string; name?: string };
    if (!body.url?.startsWith("http")) {
      return fail("INVALID_ARGUMENT", "invalid url");
    }
    const detail = addWebResource(String(params.groupId), { url: body.url, name: body.name });
    return ok({
      resource_id: detail.id,
      status: detail.status,
      job_id: detail.latest_job_id!,
    });
  }),

  http.get("/api/v1/resources/:resourceId", ({ params }) => {
    const resource = resources.find((item) => item.id === params.resourceId);
    return resource ? ok(resource) : fail("RESOURCE_NOT_FOUND", "resource not found", 404);
  }),

  http.post("/api/v1/resources/:resourceId/retry", ({ params }) => {
    const resource = retryResourceJob(String(params.resourceId));
    return resource
      ? ok({
          resource_id: resource.id,
          status: resource.status,
          job_id: resource.latest_job_id!,
        })
      : fail("RESOURCE_NOT_FOUND", "resource not found", 404);
  }),

  http.delete("/api/v1/resources/:resourceId", ({ params }) => {
    removeResource(String(params.resourceId));
    return ok({ success: true });
  }),

  http.get("/api/v1/resources/:resourceId/graph", ({ params, request }) => {
    const url = new URL(request.url);
    const includeExpansion = url.searchParams.get("include_expansion") !== "false";
    const maxLevel = Number(url.searchParams.get("max_level") || "99");
    const resource = resources.find((item) => item.id === params.resourceId);
    if (!resource?.latest_graph_id) {
      return fail("RESOURCE_NOT_FOUND", "graph not found", 404);
    }
    const graph = graphViews[resource.latest_graph_id];
    return ok({
      ...graph,
      nodes: graph.nodes.filter((node) => node.level <= maxLevel && (includeExpansion || !node.is_expansion)),
      edges: graph.edges.filter((edge) => {
        const source = graph.nodes.find((node) => node.id === edge.source_node_id);
        const target = graph.nodes.find((node) => node.id === edge.target_node_id);
        return (
          source &&
          target &&
          source.level <= maxLevel &&
          target.level <= maxLevel &&
          (includeExpansion || !edge.is_expansion_relation)
        );
      }),
    });
  }),

  http.get("/api/v1/groups/:groupId/framework-graph", ({ params }) => {
    if (params.groupId === "grp_agent") {
      return ok(graphViews.framework_agent);
    }
    return fail("RESOURCE_NOT_FOUND", "framework graph not found", 404);
  }),

  http.post("/api/v1/groups/:groupId/framework-graph/generate", ({ params }) =>
    ok({
      group_id: String(params.groupId),
      job_id: "job_framework_regenerate",
      status: "pending" as const,
    }),
  ),

  http.get("/api/v1/graphs/:graphId/nodes/:nodeId", ({ params }) => {
    const detail = nodeDetails[String(params.nodeId)];
    return detail ? ok(detail) : fail("RESOURCE_NOT_FOUND", "node not found", 404);
  }),

  http.post("/api/v1/graphs/:graphId/nodes/:nodeId/expand", ({ params }) =>
    ok({
      job_id: `job_expand_${String(params.nodeId)}`,
      status: "pending" as const,
    }),
  ),

  http.post("/api/v1/conversations", async ({ request }) => {
    const body = (await request.json()) as Conversation;
    const conversation: Conversation = {
      id: `conv_${Date.now()}`,
      group_id: body.group_id,
      graph_id: body.graph_id,
      current_node_id: body.current_node_id,
      title: body.title ?? null,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    conversations[conversation.id] = {
      conversation,
      messages: [
        {
          id: `msg_sys_${Date.now()}`,
          conversation_id: conversation.id,
          current_node_id: body.current_node_id,
          role: "system",
          content: `当前回答会围绕“${body.title ?? "当前节点"}”，并结合本分组背景知识解释。`,
          citations: { chunk_ids: [], node_ids: [] },
          created_at: new Date().toISOString(),
        },
      ],
    };
    return ok(conversation);
  }),

  http.get("/api/v1/conversations/:conversationId", ({ params }) => {
    const conversation = conversations[String(params.conversationId)];
    return conversation
      ? ok<ConversationDetail>(conversation)
      : fail("RESOURCE_NOT_FOUND", "conversation not found", 404);
  }),

  http.post("/api/v1/conversations/:conversationId/messages", async ({ params, request }) => {
    const body = (await request.json()) as { content: string };
    const conversation = conversations[String(params.conversationId)];
    if (!conversation) {
      return fail("RESOURCE_NOT_FOUND", "conversation not found", 404);
    }

    return ok<SendMessageResponse>(
      appendConversationMessage(String(params.conversationId), conversation.conversation.current_node_id, body.content),
    );
  }),

  http.get("/api/v1/jobs/:jobId", ({ params }) => {
    const job = jobs.find((item) => item.id === params.jobId);
    return job ? ok(job) : fail("RESOURCE_NOT_FOUND", "job not found", 404);
  }),
];
