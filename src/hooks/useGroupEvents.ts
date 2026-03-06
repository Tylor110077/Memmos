import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { type GroupEventName, subscribeGroupEvents } from "@/api/events";
import { queryKeys } from "@/api/queryKeys";
import type { GraphView, GroupDetail, ResourceDetail, ResourceSummary } from "@/api/types";

function patchResourceSummary(
  items: ResourceSummary[] | undefined,
  resourceId: string,
  patch: Partial<ResourceSummary>,
) {
  if (!items) return items;
  return items.map((resource) => (resource.id === resourceId ? { ...resource, ...patch } : resource));
}

export function applyGroupEventToCache(queryClient: ReturnType<typeof useQueryClient>, event: GroupEventName, payload: any) {
  if (event === "resource.status.changed") {
    queryClient.setQueriesData({ queryKey: queryKeys.resources(payload.group_id) }, (current: any) =>
      current
        ? {
            ...current,
            items: patchResourceSummary(current.items, payload.resource_id, {
              status: payload.status,
              updated_at: payload.updated_at,
            }),
          }
        : current,
    );
    queryClient.setQueriesData({ queryKey: queryKeys.resourceDetail(payload.resource_id) }, (current: ResourceDetail | undefined) =>
      current ? { ...current, status: payload.status, updated_at: payload.updated_at } : current,
    );
  }

  if (event === "resource.completed") {
    queryClient.setQueriesData({ queryKey: queryKeys.resources(payload.group_id) }, (current: any) =>
      current
        ? {
            ...current,
            items: patchResourceSummary(current.items, payload.resource_id, {
              status: payload.status,
            }),
          }
        : current,
    );
    queryClient.setQueriesData({ queryKey: queryKeys.resourceDetail(payload.resource_id) }, (current: ResourceDetail | undefined) =>
      current ? { ...current, status: payload.status, latest_graph_id: payload.graph_id } : current,
    );
    void queryClient.invalidateQueries({ queryKey: ["resources", payload.resource_id, "graph"] });
    void queryClient.invalidateQueries({ queryKey: queryKeys.groupDetail(payload.group_id) });
  }

  if (event === "resource.failed") {
    queryClient.setQueriesData({ queryKey: queryKeys.resources(payload.group_id) }, (current: any) =>
      current
        ? {
            ...current,
            items: patchResourceSummary(current.items, payload.resource_id, {
              status: payload.status,
              error_message: payload.error_message,
            }),
          }
        : current,
    );
    queryClient.setQueriesData({ queryKey: queryKeys.resourceDetail(payload.resource_id) }, (current: ResourceDetail | undefined) =>
      current ? { ...current, status: payload.status, error_message: payload.error_message } : current,
    );
  }

  if (event === "framework_graph.updated") {
    void queryClient.invalidateQueries({ queryKey: ["groups", payload.group_id, "framework-graph"] });
  }

  if (event === "graph.node.expanded") {
    void queryClient.invalidateQueries({ queryKey: ["resources"] });
    void queryClient.invalidateQueries({ queryKey: ["groups", payload.group_id, "framework-graph"] });
  }
}

export function useGroupEvents(groupId?: string) {
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!groupId) return;
    if (import.meta.env.VITE_API_MODE !== "live") return;
    if (typeof EventSource === "undefined") return;

    return subscribeGroupEvents(groupId, (event, payload) => {
      applyGroupEventToCache(queryClient, event, payload);
    });
  }, [groupId, queryClient]);
}
