import { describe, expect, it } from "vitest";
import {
  normalizeConversationDetail,
  normalizeGraphView,
  normalizeGroups,
  normalizeMessage,
  normalizeNodeDetail,
  normalizeResourceSummaries,
  normalizeSendMessageResponse,
} from "@/api/liveAdapters";

describe("liveAdapters", () => {
  it("normalizes array group responses into paginated payloads", () => {
    const result = normalizeGroups([
      {
        id: "grp_live",
        name: "Live Group",
        description: null,
        resource_count: 0,
        created_at: "2026-03-07T00:00:00Z",
      },
    ]);

    expect(result.items[0]?.id).toBe("grp_live");
    expect(result.total).toBe(1);
    expect(result.page).toBe(1);
  });

  it("normalizes resource summary aliases", () => {
    const result = normalizeResourceSummaries([
      {
        id: "res_live",
        group_id: "grp_live",
        name: "Live Resource",
        type: "html",
        status: "completed",
        error_message: null,
        created_at: "2026-03-07T00:00:00Z",
        updated_at: "2026-03-07T00:10:00Z",
      },
    ]);

    expect(result.items[0]?.resource_type).toBe("web");
  });

  it("normalizes framework graph aliases", () => {
    const result = normalizeGraphView({
      graph: {
        id: "graph_live",
        group_id: "grp_live",
        resource_id: "",
        title: "Framework",
        summary: "Overview",
        version: 1,
        is_active: true,
        created_at: "2026-03-07T00:00:00Z",
        updated_at: "2026-03-07T00:00:00Z",
      },
      nodes: [
        { id: "root", graph_id: "graph_live", name: "Framework", type: "topic", level: 0, is_expansion: false },
        { id: "n1", graph_id: "graph_live", name: "Overview", type: "theme", level: 1, is_expansion: false },
      ],
      edges: [
        {
          id: "e1",
          graph_id: "graph_live",
          source_id: "root",
          target_id: "n1",
          relation: "contains",
          is_expansion: false,
        },
      ],
    });

    expect(result.graph.graph_type).toBe("framework");
    expect(result.graph.root_node_id).toBe("root");
    expect(result.edges[0]?.source_node_id).toBe("root");
    expect(result.edges[0]?.target_node_id).toBe("n1");
  });

  it("deduplicates repeated graph nodes and edges by id", () => {
    const result = normalizeGraphView({
      graph: {
        id: "graph_live",
        group_id: "grp_live",
        resource_id: "",
        title: "Framework",
        summary: "Overview",
        version: 1,
        is_active: true,
        created_at: "2026-03-07T00:00:00Z",
        updated_at: "2026-03-07T00:00:00Z",
      },
      nodes: [
        { id: "root_exp_exp", graph_id: "graph_live", name: "Root Exp", type: "topic", level: 1, is_expansion: true },
        { id: "root_exp_exp", graph_id: "graph_live", name: "Root Exp", type: "topic", level: 1, is_expansion: true },
      ],
      edges: [
        {
          id: "root_exp_to_root_exp_exp",
          graph_id: "graph_live",
          source_id: "root_exp_exp",
          target_id: "root_exp_exp",
          relation: "extends",
          is_expansion: true,
        },
        {
          id: "root_exp_to_root_exp_exp",
          graph_id: "graph_live",
          source_id: "root_exp_exp",
          target_id: "root_exp_exp",
          relation: "extends",
          is_expansion: true,
        },
      ],
    });

    expect(result.nodes).toHaveLength(1);
    expect(result.edges).toHaveLength(1);
  });

  it("normalizes node detail aliases", () => {
    const result = normalizeNodeDetail({
      node: {
        id: "root",
        graph_id: "graph_live",
        name: "Framework",
        type: "topic",
        level: 0,
        is_expansion: false,
      },
      neighbors: [
        {
          node: {
            id: "n1",
            graph_id: "graph_live",
            name: "Overview",
            type: "theme",
            level: 1,
            is_expansion: false,
          },
          relation: "contains",
        },
      ],
      examples: [{ id: "ex1", node_id: "root", content: "Example content" }],
    });

    expect(result.level).toBe(1);
    expect(result.neighbors[0]?.node_id).toBe("n1");
    expect(result.examples[0]?.example_text).toBe("Example content");
  });

  it("normalizes conversation detail with embedded messages", () => {
    const result = normalizeConversationDetail({
      id: "conv_live",
      group_id: "grp_live",
      graph_id: "graph_live",
      current_node_id: "root",
      title: "Ask about Framework",
      created_at: "2026-03-07T00:00:00Z",
      updated_at: "2026-03-07T00:00:00Z",
      messages: [
        {
          id: "msg1",
          conversation_id: "conv_live",
          role: "assistant",
          content: "Answer",
          cited_chunk_ids: ["chunk1"],
          cited_node_ids: ["root"],
          created_at: "2026-03-07T00:01:00Z",
        },
      ],
    });

    expect(result.conversation.id).toBe("conv_live");
    expect(result.messages[0]?.citations.node_ids).toEqual(["root"]);
  });

  it("fills current node id from message context snapshot", () => {
    const result = normalizeMessage({
      id: "msg_live",
      conversation_id: "conv_live",
      role: "assistant",
      content: "Answer",
      cited_node_ids: ["root"],
      context_snapshot: {
        current_node_id: "root",
      },
      created_at: "2026-03-07T00:01:00Z",
    });

    expect(result.current_node_id).toBe("root");
    expect(result.citations.node_ids).toEqual(["root"]);
  });

  it("normalizes send message response payloads", () => {
    const result = normalizeSendMessageResponse({
      user_message: {
        id: "msg_user",
        conversation_id: "conv_live",
        role: "user",
        content: "Question",
        context_snapshot: {
          current_node_id: "root",
        },
        created_at: "2026-03-07T00:01:00Z",
      },
      assistant_message: {
        id: "msg_assistant",
        conversation_id: "conv_live",
        role: "assistant",
        content: "Answer",
        cited_chunk_ids: ["chunk1"],
        cited_node_ids: ["root"],
        context_snapshot: {
          current_node_id: "root",
        },
        created_at: "2026-03-07T00:01:05Z",
      },
    });

    expect(result.user_message.current_node_id).toBe("root");
    expect(result.assistant_message.citations.chunk_ids).toEqual(["chunk1"]);
    expect(result.assistant_message.citations.node_ids).toEqual(["root"]);
  });
});
