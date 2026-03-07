import { useMemo } from "react";
import { useQueries } from "@tanstack/react-query";
import { getResources } from "@/api/resources";
import { queryKeys } from "@/api/queryKeys";
import type { Group, ResourceSummary } from "@/api/types";

export function useGroupResourceIndex(groups: Group[]) {
  const resourceQueries = useQueries({
    queries: groups.map((group) => ({
      queryKey: queryKeys.resources(group.id),
      queryFn: () => getResources(group.id),
      enabled: Boolean(group.id),
      staleTime: 60_000,
    })),
  });

  const resourcesByGroupId = useMemo(
    () =>
      groups.reduce<Record<string, ResourceSummary[]>>((accumulator, group, index) => {
        accumulator[group.id] = resourceQueries[index]?.data?.items ?? [];
        return accumulator;
      }, {}),
    [groups, resourceQueries],
  );

  const resources = useMemo(() => Object.values(resourcesByGroupId).flat(), [resourcesByGroupId]);

  return {
    resourcesByGroupId,
    resources,
    isLoading: resourceQueries.some((query) => query.isLoading),
  };
}
