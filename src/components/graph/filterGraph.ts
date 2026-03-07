import type { GraphView, NodeType } from "@/api/types";

type FilterOptions = {
  searchTerm: string;
  nodeTypeFilter: NodeType[];
};

export function filterGraphView(graph: GraphView, options: FilterOptions): GraphView {
  const normalizedSearch = options.searchTerm.trim().toLowerCase();
  const hasTypeFilter = options.nodeTypeFilter.length > 0;

  const nodes = graph.nodes.filter((node) => {
    const matchesSearch =
      normalizedSearch.length === 0 ||
      node.name.toLowerCase().includes(normalizedSearch) ||
      node.description?.toLowerCase().includes(normalizedSearch) ||
      node.meaning?.toLowerCase().includes(normalizedSearch);
    const matchesType = !hasTypeFilter || options.nodeTypeFilter.includes(node.node_type);

    return matchesSearch && matchesType;
  });

  const visibleIds = new Set(nodes.map((node) => node.id));
  const edges = graph.edges.filter((edge) => visibleIds.has(edge.source_node_id) && visibleIds.has(edge.target_node_id));
  const nextRootId = visibleIds.has(graph.graph.root_node_id ?? "")
    ? graph.graph.root_node_id
    : nodes[0]?.id ?? null;

  return {
    graph: {
      ...graph.graph,
      root_node_id: nextRootId,
    },
    nodes,
    edges,
  };
}
