const API_BASE = process.env.LIVE_API_BASE ?? "http://127.0.0.1:8080/api/v1";

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      Accept: "application/json",
      ...(options.body instanceof FormData ? {} : { "Content-Type": "application/json" }),
      ...options.headers,
    },
  });

  const payload = await response.json();
  if (!response.ok || payload.error) {
    throw new Error(`${options.method ?? "GET"} ${path} failed: ${payload.error?.message ?? response.status}`);
  }
  return payload.data;
}

async function verifyGroupEventsSse(groupId) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 1500);

  try {
    const response = await fetch(`${API_BASE}/events/groups/${groupId}`, {
      headers: {
        Accept: "text/event-stream",
      },
      signal: controller.signal,
    });
    assert(response.ok, "group events SSE request failed");
    assert(response.headers.get("content-type")?.includes("text/event-stream"), "group events SSE content-type mismatch");
  } catch (error) {
    if (error.name !== "AbortError") {
      throw error;
    }
  } finally {
    clearTimeout(timeout);
  }
}

async function readStream(response) {
  const reader = response.body?.getReader();
  assert(reader, "stream reader unavailable");

  const decoder = new TextDecoder("utf-8");
  let buffer = "";
  const events = [];

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const chunks = buffer.split("\n\n");
    buffer = chunks.pop() ?? "";

    for (const chunk of chunks) {
      const lines = chunk.split("\n");
      const event = lines.find((line) => line.startsWith("event:"))?.slice(6).trim();
      const data = lines.find((line) => line.startsWith("data:"))?.slice(5).trim();
      if (!event || !data) continue;
      events.push({ event, data: JSON.parse(data) });
    }
  }

  return events;
}

async function verifyConversationStream(conversationId) {
  const response = await fetch(`${API_BASE}/conversations/${conversationId}/messages`, {
    method: "POST",
    headers: {
      Accept: "text/event-stream",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ content: "stream smoke", stream: true }),
  });

  assert(response.ok, "stream message request failed");
  assert(response.headers.get("content-type")?.includes("text/event-stream"), "stream message content-type mismatch");

  const events = await readStream(response);
  const eventNames = events.map((item) => item.event);

  assert(eventNames.includes("assistant.delta"), "missing assistant.delta event");
  assert(eventNames.includes("assistant.message"), "missing assistant.message event");
  assert(eventNames.includes("done"), "missing done event");
}

async function main() {
  const ready = await fetch("http://127.0.0.1:8080/readyz");
  assert(ready.ok, "readyz is not available");

  const group = await request("/groups", {
    method: "POST",
    body: JSON.stringify({
      name: "Live Smoke Group",
      description: "final acceptance smoke test",
    }),
  });

  const resourceCreate = await request(`/groups/${group.id}/web-resources`, {
    method: "POST",
    body: JSON.stringify({
      url: "https://example.com/live-smoke",
      name: "Live Smoke Resource",
    }),
  });

  const resourceId = resourceCreate.resource?.id ?? resourceCreate.resource_id;
  assert(resourceId, "resource id missing from web resource response");

  const groupDetail = await request(`/groups/${group.id}`);
  assert(groupDetail.resource_count >= 1, "group resource_count did not update");

  const resources = await request(`/groups/${group.id}/resources`);
  assert(Array.isArray(resources) && resources.some((resource) => resource.id === resourceId), "resource list missing created resource");

  const graph = await request(`/resources/${resourceId}/graph`);
  assert(graph.graph?.id, "resource graph missing graph id");
  assert(Array.isArray(graph.nodes) && graph.nodes.length > 0, "resource graph missing nodes");

  const conversation = await request("/conversations", {
    method: "POST",
    body: JSON.stringify({
      group_id: group.id,
      graph_id: graph.graph.id,
      current_node_id: graph.graph.root_node_id ?? "root",
      title: "Live Smoke Conversation",
    }),
  });

  const messageResult = await request(`/conversations/${conversation.id}/messages`, {
    method: "POST",
    body: JSON.stringify({
      content: "请解释这个资源图谱",
      stream: false,
    }),
  });
  assert(messageResult.assistant_message?.id, "assistant message missing in non-stream response");

  const conversationDetail = await request(`/conversations/${conversation.id}`);
  assert(Array.isArray(conversationDetail.messages) && conversationDetail.messages.length >= 2, "conversation detail missing messages");

  await verifyConversationStream(conversation.id);
  await verifyGroupEventsSse(group.id);

  console.log(JSON.stringify({
    ok: true,
    group_id: group.id,
    resource_id: resourceId,
    graph_id: graph.graph.id,
    conversation_id: conversation.id,
  }));
}

main().catch((error) => {
  console.error(error instanceof Error ? error.message : String(error));
  process.exit(1);
});
