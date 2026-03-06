import { apiRequest } from "@/api/client";
import type {
  Conversation,
  ConversationDetail,
  ConversationMessage,
  SendMessageResponse,
} from "@/api/types";

export function createConversation(payload: {
  group_id: string;
  graph_id: string;
  current_node_id: string;
  title?: string;
}) {
  return apiRequest<Conversation>("/conversations", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function getConversation(conversationId: string) {
  return apiRequest<ConversationDetail>(`/conversations/${conversationId}`);
}

export function sendConversationMessage(conversationId: string, content: string, stream = false) {
  return apiRequest<SendMessageResponse>(`/conversations/${conversationId}/messages`, {
    method: "POST",
    body: JSON.stringify({ content, stream }),
  });
}

export type StreamConversationHandlers = {
  onStart?: (messageId: string) => void;
  onDelta?: (delta: string) => void;
  onDone?: (message: ConversationMessage) => void;
};

export async function streamConversationMessage(
  conversationId: string,
  content: string,
  handlers: StreamConversationHandlers = {},
) {
  const response = await fetch(`/api/v1/conversations/${conversationId}/messages`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "text/event-stream",
    },
    body: JSON.stringify({ content, stream: true }),
  });

  if (!response.ok) {
    throw new Error("stream request failed");
  }

  const reader = response.body?.getReader();
  if (!reader) {
    throw new Error("stream reader unavailable");
  }

  const decoder = new TextDecoder("utf-8");
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
