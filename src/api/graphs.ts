import { apiRequest } from "@/api/client";
import type {
  ExpandNodeResult,
  FrameworkGraphGenerateResult,
  GraphView,
  NodeDetail,
} from "@/api/types";
import { normalizeGraphView, normalizeNodeDetail } from "@/api/liveAdapters";

export async function getResourceGraph(resourceId: string, params: { includeExpansion: boolean; maxLevel?: number }) {
  const query = new URLSearchParams();
  query.set("include_expansion", String(params.includeExpansion));
  if (params.maxLevel) {
    query.set("max_level", String(params.maxLevel));
  }

  const response = await apiRequest<GraphView>(`/resources/${resourceId}/graph?${query.toString()}`);
  return normalizeGraphView(response);
}

export async function getFrameworkGraph(groupId: string, maxLevel = 2) {
  const response = await apiRequest<GraphView>(`/groups/${groupId}/framework-graph?max_level=${maxLevel}`);
  return normalizeGraphView(response);
}

export function generateFrameworkGraph(groupId: string) {
  return apiRequest<FrameworkGraphGenerateResult>(`/groups/${groupId}/framework-graph/generate`, {
    method: "POST",
    body: JSON.stringify({}),
  });
}

export async function getNodeDetail(graphId: string, nodeId: string) {
  const response = await apiRequest<NodeDetail>(`/graphs/${graphId}/nodes/${nodeId}`);
  return normalizeNodeDetail(response);
}

export function expandNode(graphId: string, nodeId: string, reason: string) {
  return apiRequest<ExpandNodeResult>(`/graphs/${graphId}/nodes/${nodeId}/expand`, {
    method: "POST",
    body: JSON.stringify({ reason }),
  });
}
