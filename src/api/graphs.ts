import { apiRequest } from "@/api/client";
import type {
  ExpandNodeResult,
  FrameworkGraphGenerateResult,
  GraphView,
  NodeDetail,
} from "@/api/types";

export function getResourceGraph(resourceId: string, params: { includeExpansion: boolean; maxLevel?: number }) {
  const query = new URLSearchParams();
  query.set("include_expansion", String(params.includeExpansion));
  if (params.maxLevel) {
    query.set("max_level", String(params.maxLevel));
  }

  return apiRequest<GraphView>(`/resources/${resourceId}/graph?${query.toString()}`);
}

export function getFrameworkGraph(groupId: string, maxLevel = 2) {
  return apiRequest<GraphView>(`/groups/${groupId}/framework-graph?max_level=${maxLevel}`);
}

export function generateFrameworkGraph(groupId: string) {
  return apiRequest<FrameworkGraphGenerateResult>(`/groups/${groupId}/framework-graph/generate`, {
    method: "POST",
    body: JSON.stringify({}),
  });
}

export function getNodeDetail(graphId: string, nodeId: string) {
  return apiRequest<NodeDetail>(`/graphs/${graphId}/nodes/${nodeId}`);
}

export function expandNode(graphId: string, nodeId: string, reason: string) {
  return apiRequest<ExpandNodeResult>(`/graphs/${graphId}/nodes/${nodeId}/expand`, {
    method: "POST",
    body: JSON.stringify({ reason }),
  });
}
