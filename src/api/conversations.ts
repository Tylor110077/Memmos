import { apiRequest } from "@/api/client";
import type {
  Conversation,
  ConversationDetail,
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

export function sendConversationMessage(conversationId: string, content: string) {
  return apiRequest<SendMessageResponse>(`/conversations/${conversationId}/messages`, {
    method: "POST",
    body: JSON.stringify({ content, stream: false }),
  });
}
