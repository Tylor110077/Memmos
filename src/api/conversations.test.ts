import { afterEach, describe, expect, it, vi } from "vitest";
import { streamConversationMessage } from "@/api/conversations";
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
});
