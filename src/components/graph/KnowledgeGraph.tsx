import { useEffect, useMemo, useState } from "react";
import ELK from "elkjs/lib/elk.bundled.js";
import {
  Background,
  Controls,
  MarkerType,
  MiniMap,
  Panel,
  ReactFlow,
  type Edge,
  type Node,
  type ReactFlowInstance,
  useEdgesState,
  useNodesState,
} from "@xyflow/react";
import { useTheme } from "@/components/theme/ThemeProvider";
import { Button } from "@/components/ui/Button";
import type { GraphView } from "@/api/types";
import { cn } from "@/lib/utils";
import { buildGraphSvg, downloadGraphSvg } from "@/components/graph/exportGraph";

const elk = new ELK();

type KnowledgeGraphProps = {
  graph: GraphView;
  selectedNodeId?: string;
  onSelectNode?: (nodeId: string) => void;
  variant?: "default" | "preview";
  exportName?: string;
};

async function layoutGraph(graph: GraphView, variant: "default" | "preview") {
  const isPreview = variant === "preview";
  const layout = await elk.layout({
    id: "root",
    layoutOptions: {
      "elk.algorithm": "layered",
      "elk.direction": "DOWN",
      "elk.layered.spacing.nodeNodeBetweenLayers": isPreview ? "28" : "60",
      "elk.spacing.nodeNode": isPreview ? "18" : "32",
    },
    children: graph.nodes.map((node) => ({
      id: node.id,
      width: isPreview ? (node.level === 1 ? 132 : 104) : node.level === 1 ? 220 : 180,
      height: isPreview ? 48 : node.level === 3 ? 64 : 72,
    })),
    edges: graph.edges.map((edge) => ({
      id: edge.id,
      sources: [edge.source_node_id],
      targets: [edge.target_node_id],
    })),
  });

  return graph.nodes.map((node) => {
    const layoutNode = layout.children?.find((item) => item.id === node.id);
    return {
      id: node.id,
      position: {
        x: layoutNode?.x ?? 0,
        y: layoutNode?.y ?? 0,
      },
      data: { label: node.name, level: node.level, isExpansion: node.is_expansion },
      type: "default",
    } satisfies Node;
  });
}

export function KnowledgeGraph({
  graph,
  selectedNodeId,
  onSelectNode,
  variant = "default",
  exportName,
}: KnowledgeGraphProps) {
  const { theme } = useTheme();
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [ready, setReady] = useState(false);
  const [flowInstance, setFlowInstance] = useState<ReactFlowInstance<Node, Edge> | null>(null);

  useEffect(() => {
    let active = true;
    void layoutGraph(graph, variant).then((layoutedNodes) => {
      if (!active) return;
      setNodes(layoutedNodes);
      setEdges(
        graph.edges.map((edge) => ({
          id: edge.id,
          source: edge.source_node_id,
          target: edge.target_node_id,
          animated: edge.is_expansion_relation,
          markerEnd: {
            type: MarkerType.ArrowClosed,
            width: 18,
            height: 18,
          },
          style: edge.is_expansion_relation
            ? { stroke: "var(--extend-line)" }
            : { stroke: "rgba(123, 151, 205, 0.58)" },
        })) satisfies Edge[],
      );
      setReady(true);
    });
    return () => {
      active = false;
      setReady(false);
    };
  }, [graph, setEdges, setNodes, variant]);

  useEffect(() => {
    if (!flowInstance || !ready || !nodes.length) return;

    requestAnimationFrame(() => {
      flowInstance.fitView({
        padding: variant === "preview" ? 0.18 : 0.18,
        minZoom: variant === "preview" ? 0.32 : 0.35,
        maxZoom: variant === "preview" ? 1.1 : 1.2,
        duration: 180,
      });
    });
  }, [flowInstance, nodes.length, ready, variant]);

  const nodeClassName = useMemo(
    () =>
      nodes.reduce<Record<string, string>>((accumulator, node) => {
        const data = node.data as { level: number; isExpansion: boolean };
        accumulator[node.id] = cn(
          "rf-node-card",
          variant === "preview" && "preview",
          data.level === 1 && "level-root",
          data.level === 2 && "level-2",
          data.level >= 3 && "level-3",
          data.isExpansion && "level-extension",
          selectedNodeId === node.id && "selected",
        );
        return accumulator;
      }, {}),
    [nodes, selectedNodeId, variant],
  );

  function handleExport() {
    const svg = buildGraphSvg({
      graph,
      nodes,
      edges,
      title: exportName ?? graph.graph.summary ?? "knowledge-graph",
      theme,
    });
    downloadGraphSvg(exportName ?? graph.graph.summary ?? "knowledge-graph", svg);
  }

  return (
    <div className={cn("graph-flow-shell", variant === "preview" && "graph-flow-shell-preview")} data-ready={ready}>
      <ReactFlow
        nodes={nodes.map((node) => ({
          ...node,
          className: nodeClassName[node.id],
        }))}
        edges={edges}
        proOptions={{ hideAttribution: true }}
        onInit={setFlowInstance}
        nodesDraggable={variant !== "preview"}
        nodesConnectable={false}
        elementsSelectable={variant !== "preview"}
        zoomOnScroll={variant !== "preview"}
        panOnDrag={variant !== "preview"}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={(_, node) => onSelectNode?.(node.id)}
        fitView
      >
        <Background color="rgba(148, 163, 184, 0.22)" gap={28} />
        {variant === "default" ? <MiniMap pannable zoomable /> : null}
        {variant === "default" ? <Controls /> : null}
        {variant === "default" ? (
          <Panel position="top-left">
            <div className="graph-panel-tag">{graph.graph.summary ?? "知识图谱"}</div>
          </Panel>
        ) : null}
        {variant === "default" ? (
          <Panel position="top-right">
            <Button variant="ghost" onClick={handleExport} disabled={!ready || !nodes.length}>
              导出 SVG
            </Button>
          </Panel>
        ) : null}
      </ReactFlow>
    </div>
  );
}
