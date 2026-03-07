import type { Edge, Node } from "@xyflow/react";
import { describe, expect, it } from "vitest";
import { buildGraphSvg } from "@/components/graph/exportGraph";
import type { GraphView } from "@/api/types";

describe("buildGraphSvg", () => {
  it("renders a standalone svg with graph title, nodes and edges", () => {
    const graph: GraphView = {
      graph: {
        id: "graph_1",
        group_id: "grp_1",
        resource_id: null,
        graph_type: "framework",
        status: "active",
        version: 1,
        summary: "知识骨架",
        root_node_id: "root",
        created_at: "2026-03-07T00:00:00Z",
        updated_at: "2026-03-07T00:00:00Z",
      },
      nodes: [],
      edges: [],
    };

    const nodes: Node[] = [
      {
        id: "root",
        position: { x: 0, y: 0 },
        width: 180,
        height: 72,
        data: { label: "Framework", level: 1, isExpansion: false },
      },
      {
        id: "child",
        position: { x: 0, y: 140 },
        width: 180,
        height: 72,
        data: { label: "Overview", level: 2, isExpansion: false },
      },
    ];
    const edges: Edge[] = [
      {
        id: "edge_1",
        source: "root",
        target: "child",
      },
    ];

    const svg = buildGraphSvg({
      graph,
      nodes,
      edges,
      title: "framework export",
      theme: "dark",
    });

    expect(svg).toContain("<svg");
    expect(svg).toContain("framework export");
    expect(svg).toContain("Framework");
    expect(svg).toContain("Overview");
    expect(svg).toContain("<path");
    expect(svg).toContain("#0f172a");
  });
});
