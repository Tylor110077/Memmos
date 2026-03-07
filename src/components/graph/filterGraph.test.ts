import { describe, expect, it } from "vitest";
import { filterGraphView } from "@/components/graph/filterGraph";
import type { GraphView } from "@/api/types";

const graph: GraphView = {
  graph: {
    id: "graph_1",
    group_id: "grp_1",
    resource_id: null,
    graph_type: "framework",
    status: "active",
    version: 1,
    summary: "Graph summary",
    root_node_id: "root",
    created_at: "2026-03-07T00:00:00Z",
    updated_at: "2026-03-07T00:00:00Z",
  },
  nodes: [
    {
      id: "root",
      graph_id: "graph_1",
      name: "Agent Framework",
      description: "Overview",
      meaning: null,
      level: 1,
      node_type: "topic",
      source_type: "summarized",
      is_expansion: false,
    },
    {
      id: "n1",
      graph_id: "graph_1",
      name: "Planning Loop",
      description: "Plan then act",
      meaning: null,
      level: 2,
      node_type: "method",
      source_type: "summarized",
      is_expansion: false,
    },
    {
      id: "n2",
      graph_id: "graph_1",
      name: "Worked Example",
      description: "Example path",
      meaning: null,
      level: 2,
      node_type: "example",
      source_type: "summarized",
      is_expansion: false,
    },
  ],
  edges: [
    {
      id: "e1",
      graph_id: "graph_1",
      source_node_id: "root",
      target_node_id: "n1",
      relation_type: "contains",
      relation_description: null,
      is_expansion_relation: false,
    },
    {
      id: "e2",
      graph_id: "graph_1",
      source_node_id: "root",
      target_node_id: "n2",
      relation_type: "contains",
      relation_description: null,
      is_expansion_relation: false,
    },
  ],
};

describe("filterGraphView", () => {
  it("filters nodes by search term", () => {
    const result = filterGraphView(graph, {
      searchTerm: "planning",
      nodeTypeFilter: [],
    });

    expect(result.nodes.map((node) => node.id)).toEqual(["n1"]);
    expect(result.graph.root_node_id).toBe("n1");
  });

  it("filters nodes by multiple node types", () => {
    const result = filterGraphView(graph, {
      searchTerm: "",
      nodeTypeFilter: ["topic", "example"],
    });

    expect(result.nodes.map((node) => node.id)).toEqual(["root", "n2"]);
    expect(result.edges.map((edge) => edge.id)).toEqual(["e2"]);
  });
});
