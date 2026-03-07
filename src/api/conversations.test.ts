import { afterEach, describe, expect, it, vi } from "vitest";
import { sendConversationMessage, streamConversationMessage } from "@/api/conversations";
import type { ConversationMessage } from "@/api/types";

function createStreamResponse(chunks: string[]) {
  return new Response(
    new ReadableStream({
      start(controller) {
        chunks.forEach((chunk) => controller.enqueue(new TextEncoder().encode(chunk)));
        controller.close();
      },
    }),
    {
      status: 200,
      headers: {
        "Content-Type": "text/event-stream",
      },
    },
  );
}

describe("streamConversationMessage", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("parses start, delta and done events in order", async () => {
    const message: ConversationMessage = {
      id: "msg_a_001",
      conversation_id: "conv_001",
      current_node_id: "node_001",
      role: "assistant",
      content: "先拆目标，再调用工具。",
      citations: {
        chunk_ids: ["chunk_001"],
        node_ids: ["node_002"],
      },
      created_at: "2026-03-06T12:32:00Z",
    };
    const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      createStreamResponse([
        'event: message.start\ndata: {"message_id":"msg_a_001"}\n\n',
        'event: message.delta\ndata: {"delta":"先拆目标，"}\n\n',
        'event: message.delta\ndata: {"delta":"再调用工具。"}\n\n',
        `event: message.done\ndata: ${JSON.stringify({ message })}\n\n`,
      ]),
    );
    const handlers = {
      onStart: vi.fn(),
      onDelta: vi.fn(),
      onDone: vi.fn(),
    };

    await streamConversationMessage("conv_001", "解释一下", handlers);

    expect(fetchSpy).toHaveBeenCalledWith("/api/v1/conversations/conv_001/messages", expect.objectContaining({
      method: "POST",
      headers: expect.objectContaining({
        Accept: "text/event-stream",
        "Content-Type": "application/json",
      }),
    }));
    expect(handlers.onStart).toHaveBeenCalledWith("msg_a_001");
    expect(handlers.onDelta).toHaveBeenNthCalledWith(1, "先拆目标，");
    expect(handlers.onDelta).toHaveBeenNthCalledWith(2, "再调用工具。");
    expect(handlers.onDone).toHaveBeenCalledWith(message);
  });

  it("normalizes non-stream message responses", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            user_message: {
              id: "msg_user_001",
              conversation_id: "conv_001",
              role: "user",
              content: "解释一下",
              context_snapshot: {
                current_node_id: "root",
              },
              created_at: "2026-03-06T12:30:00Z",
            },
            assistant_message: {
              id: "msg_assistant_001",
              conversation_id: "conv_001",
              role: "assistant",
              content: "先拆目标，再调用工具。",
              cited_chunk_ids: ["chunk_001"],
              cited_node_ids: ["root"],
              context_snapshot: {
                current_node_id: "root",
              },
              created_at: "2026-03-06T12:32:00Z",
            },
          },
          error: null,
          meta: {},
        }),
        {
          status: 200,
          headers: {
            "Content-Type": "application/json",
          },
        },
      ),
    );

    const response = await sendConversationMessage("conv_001", "解释一下");

    expect(response.user_message.current_node_id).toBe("root");
    expect(response.assistant_message.citations.chunk_ids).toEqual(["chunk_001"]);
    expect(response.assistant_message.citations.node_ids).toEqual(["root"]);
  });

  it("supports backend assistant.delta and assistant.message stream events", async () => {
    const doneMessage: ConversationMessage = {
      id: "msg_assistant_002",
      conversation_id: "conv_001",
      current_node_id: "root",
      role: "assistant",
      content: "Graph Ready Resource is the current focus node.",
      citations: {
        chunk_ids: [],
        node_ids: ["root"],
      },
      created_at: "2026-03-06T12:33:00Z",
    };
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      createStreamResponse([
        'event: assistant.delta\ndata: {"delta":"Graph Ready Resource "}\n\n',
        'event: assistant.delta\ndata: {"delta":"is the current focus node."}\n\n',
        `event: assistant.message\ndata: ${JSON.stringify({
          assistant_message: {
            id: "msg_assistant_002",
            conversation_id: "conv_001",
            role: "assistant",
            content: "Graph Ready Resource is the current focus node.",
            cited_chunk_ids: [],
            cited_node_ids: ["root"],
            context_snapshot: { current_node_id: "root" },
            created_at: "2026-03-06T12:33:00Z",
          },
        })}\n\n`,
        'event: done\ndata: {"done":true}\n\n',
      ]),
    );

    const handlers = {
      onStart: vi.fn(),
      onDelta: vi.fn(),
      onDone: vi.fn(),
    };

    await streamConversationMessage("conv_001", "解释一下", handlers);

    expect(handlers.onStart).not.toHaveBeenCalled();
    expect(handlers.onDelta).toHaveBeenNthCalledWith(1, "Graph Ready Resource ");
    expect(handlers.onDelta).toHaveBeenNthCalledWith(2, "is the current focus node.");
    expect(handlers.onDone).toHaveBeenCalledWith(doneMessage);
  });
});
