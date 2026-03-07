import type { Edge, Node } from "@xyflow/react";
import type { GraphView } from "@/api/types";

type ExportTheme = "light" | "dark";

const palettes = {
  light: {
    background: "#f7fafd",
    border: "#d7dee7",
    edge: "#8aa4d3",
    text: "#1f2937",
    muted: "#5b6b83",
    root: { fill: "#dce8ff", stroke: "#8fb0ff", text: "#23408e" },
    level2: { fill: "#e9f1ff", stroke: "#b3c8ff", text: "#304b8b" },
    level3: { fill: "#f3f7ff", stroke: "#cbd9f6", text: "#42526e" },
    extension: { fill: "#f5f3ff", stroke: "#cfc4ff", text: "#6d5bb3" },
  },
  dark: {
    background: "#0f172a",
    border: "#22314a",
    edge: "#5678b3",
    text: "#e8eefb",
    muted: "#9fb0cf",
    root: { fill: "#18284b", stroke: "#5f88df", text: "#dbe8ff" },
    level2: { fill: "#16233d", stroke: "#496592", text: "#d5e1fa" },
    level3: { fill: "#132038", stroke: "#32496c", text: "#cfdaef" },
    extension: { fill: "#221735", stroke: "#604692", text: "#d2c2ff" },
  },
} satisfies Record<ExportTheme, Record<string, unknown>>;

function escapeXml(value: string) {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&apos;");
}

function sanitizeFileName(value: string) {
  return value.replace(/[^\w.-]+/g, "-").replace(/^-+|-+$/g, "").toLowerCase() || "knowledge-graph";
}

function getNodePalette(level: number, isExpansion: boolean, theme: ExportTheme) {
  const palette = palettes[theme];
  if (isExpansion) return palette.extension;
  if (level === 1) return palette.root;
  if (level === 2) return palette.level2;
  return palette.level3;
}

function linePath(source: Node, target: Node) {
  const sourceX = source.position.x + ((source.width as number | undefined) ?? 180) / 2;
  const sourceY = source.position.y + ((source.height as number | undefined) ?? 72);
  const targetX = target.position.x + ((target.width as number | undefined) ?? 180) / 2;
  const targetY = target.position.y;
  const controlOffset = Math.max(32, Math.abs(targetY - sourceY) * 0.35);

  return `M ${sourceX} ${sourceY} C ${sourceX} ${sourceY + controlOffset}, ${targetX} ${targetY - controlOffset}, ${targetX} ${targetY}`;
}

export function buildGraphSvg({
  graph,
  nodes,
  edges,
  title,
  theme,
}: {
  graph: GraphView;
  nodes: Node[];
  edges: Edge[];
  title: string;
  theme: ExportTheme;
}) {
  const visibleNodes = nodes.filter((node) => Number.isFinite(node.position.x) && Number.isFinite(node.position.y));
  if (!visibleNodes.length) {
    throw new Error("graph has no positioned nodes");
  }

  const padding = 32;
  const minX = Math.min(...visibleNodes.map((node) => node.position.x));
  const minY = Math.min(...visibleNodes.map((node) => node.position.y));
  const maxX = Math.max(...visibleNodes.map((node) => node.position.x + (((node.width as number | undefined) ?? 180))));
  const maxY = Math.max(...visibleNodes.map((node) => node.position.y + (((node.height as number | undefined) ?? 72))));
  const width = Math.max(720, Math.ceil(maxX - minX + padding * 2));
  const height = Math.max(480, Math.ceil(maxY - minY + padding * 2 + 56));
  const offsetX = padding - minX;
  const offsetY = padding + 36 - minY;
  const palette = palettes[theme];

  const svgEdges = edges
    .map((edge) => {
      const source = visibleNodes.find((node) => node.id === edge.source);
      const target = visibleNodes.find((node) => node.id === edge.target);
      if (!source || !target) return "";
      const d = linePath(
        { ...source, position: { x: source.position.x + offsetX, y: source.position.y + offsetY } },
        { ...target, position: { x: target.position.x + offsetX, y: target.position.y + offsetY } },
      );
      return `<path d="${d}" fill="none" stroke="${palette.edge}" stroke-width="2.4" stroke-linecap="round" marker-end="url(#arrow)" />`;
    })
    .join("");

  const svgNodes = visibleNodes
    .map((node) => {
      const data = node.data as { label: string; level: number; isExpansion: boolean };
      const width = ((node.width as number | undefined) ?? 180);
      const height = ((node.height as number | undefined) ?? 72);
      const x = node.position.x + offsetX;
      const y = node.position.y + offsetY;
      const palette = getNodePalette(data.level, data.isExpansion, theme);
      const textY = y + height / 2 + 5;

      return [
        `<rect x="${x}" y="${y}" width="${width}" height="${height}" rx="16" fill="${palette.fill}" stroke="${palette.stroke}" stroke-width="1.5" />`,
        `<text x="${x + width / 2}" y="${textY}" text-anchor="middle" font-family="Inter, Noto Sans SC, sans-serif" font-size="14" font-weight="600" fill="${palette.text}">${escapeXml(data.label)}</text>`,
      ].join("");
    })
    .join("");

  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" role="img" aria-label="${escapeXml(title)}">
  <defs>
    <marker id="arrow" markerWidth="10" markerHeight="10" refX="8" refY="3" orient="auto" markerUnits="strokeWidth">
      <path d="M0,0 L0,6 L9,3 z" fill="${palette.edge}" />
    </marker>
  </defs>
  <rect width="${width}" height="${height}" rx="24" fill="${palette.background}" stroke="${palette.border}" />
  <text x="32" y="34" font-family="Inter, Noto Sans SC, sans-serif" font-size="20" font-weight="700" fill="${palette.text}">${escapeXml(title)}</text>
  <text x="32" y="58" font-family="Inter, Noto Sans SC, sans-serif" font-size="12" fill="${palette.muted}">${escapeXml(graph.graph.summary ?? "知识图谱导出")}</text>
  ${svgEdges}
  ${svgNodes}
</svg>`;
}

export function downloadGraphSvg(fileName: string, svg: string) {
  const blob = new Blob([svg], { type: "image/svg+xml;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = `${sanitizeFileName(fileName)}.svg`;
  document.body.append(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
}
