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
  useEdgesState,
  useNodesState,
} from "@xyflow/react";
import type { GraphView } from "@/api/types";
import { cn } from "@/lib/utils";

const elk = new ELK();

type KnowledgeGraphProps = {
  graph: GraphView;
  selectedNodeId?: string;
  onSelectNode?: (nodeId: string) => void;
  variant?: "default" | "preview";
};

async function layoutGraph(graph: GraphView) {
  const layout = await elk.layout({
    id: "root",
    layoutOptions: {
      "elk.algorithm": "layered",
      "elk.direction": "DOWN",
      "elk.layered.spacing.nodeNodeBetweenLayers": "60",
      "elk.spacing.nodeNode": "32",
    },
    children: graph.nodes.map((node) => ({
      id: node.id,
      width: node.level === 1 ? 220 : 180,
      height: node.level === 3 ? 64 : 72,
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
}: KnowledgeGraphProps) {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    let active = true;
    void layoutGraph(graph).then((layoutedNodes) => {
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
  }, [graph, setEdges, setNodes]);

  const nodeClassName = useMemo(
    () =>
      nodes.reduce<Record<string, string>>((accumulator, node) => {
        const data = node.data as { level: number; isExpansion: boolean };
        accumulator[node.id] = cn(
          "rf-node-card",
          data.level === 1 && "level-root",
          data.level === 2 && "level-2",
          data.level >= 3 && "level-3",
          data.isExpansion && "level-extension",
          selectedNodeId === node.id && "selected",
        );
        return accumulator;
      }, {}),
    [nodes, selectedNodeId],
  );

  return (
    <div className={cn("graph-flow-shell", variant === "preview" && "graph-flow-shell-preview")} data-ready={ready}>
      <ReactFlow
        nodes={nodes.map((node) => ({
          ...node,
          className: nodeClassName[node.id],
        }))}
        edges={edges}
        proOptions={{ hideAttribution: true }}
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
        <Panel position="top-left">
          <div className="graph-panel-tag">{graph.graph.summary ?? "知识图谱"}</div>
        </Panel>
      </ReactFlow>
    </div>
  );
}
