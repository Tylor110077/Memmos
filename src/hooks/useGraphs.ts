import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { expandNode, generateFrameworkGraph, getFrameworkGraph, getNodeDetail, getResourceGraph } from "@/api/graphs";
import { queryKeys } from "@/api/queryKeys";

export function useFrameworkGraph(groupId: string, maxLevel = 2) {
  return useQuery({
    queryKey: queryKeys.frameworkGraph(groupId, maxLevel),
    queryFn: () => getFrameworkGraph(groupId, maxLevel),
    enabled: Boolean(groupId),
  });
}

export function useGenerateFrameworkGraph(groupId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => generateFrameworkGraph(groupId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["groups", groupId, "framework-graph"] });
    },
  });
}

export function useResourceGraph(resourceId: string, params: { includeExpansion: boolean; maxLevel?: number }) {
  return useQuery({
    queryKey: queryKeys.resourceGraph(resourceId, params),
    queryFn: () => getResourceGraph(resourceId, params),
    enabled: Boolean(resourceId),
  });
}

export function useNodeDetail(graphId: string, nodeId?: string) {
  return useQuery({
    queryKey: queryKeys.nodeDetail(graphId, nodeId ?? ""),
    queryFn: () => getNodeDetail(graphId, nodeId ?? ""),
    enabled: Boolean(graphId && nodeId),
  });
}

export function useExpandNode(graphId: string, resourceId: string, groupId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ nodeId, reason }: { nodeId: string; reason: string }) => expandNode(graphId, nodeId, reason),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["resources", resourceId, "graph"] });
      void queryClient.invalidateQueries({ queryKey: ["groups", groupId, "framework-graph"] });
    },
  });
}
