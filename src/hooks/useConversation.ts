import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createConversation,
  getConversation,
  sendConversationMessage,
  streamConversationMessage,
  type StreamConversationHandlers,
} from "@/api/conversations";
import { queryKeys } from "@/api/queryKeys";

export function useConversation(conversationId?: string) {
  return useQuery({
    queryKey: queryKeys.conversation(conversationId ?? ""),
    queryFn: () => getConversation(conversationId ?? ""),
    enabled: Boolean(conversationId),
  });
}

export function useCreateConversation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createConversation,
    onSuccess: (conversation) => {
      queryClient.setQueryData(queryKeys.conversation(conversation.id), {
        conversation,
        messages: [],
      });
    },
  });
}

export function useSendMessage(conversationId?: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (content: string) => sendConversationMessage(conversationId ?? "", content),
    onSuccess: async () => {
      if (conversationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.conversation(conversationId) });
      }
    },
  });
}

export function useStreamMessage(conversationId?: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      conversationId: targetConversationId,
      content,
      handlers,
    }: {
      conversationId?: string;
      content: string;
      handlers?: StreamConversationHandlers;
    }) => streamConversationMessage(targetConversationId ?? conversationId ?? "", content, handlers),
    onSuccess: async (_, variables) => {
      const targetConversationId = variables.conversationId ?? conversationId;
      if (targetConversationId) {
        await queryClient.invalidateQueries({ queryKey: queryKeys.conversation(targetConversationId) });
      }
    },
  });
}
