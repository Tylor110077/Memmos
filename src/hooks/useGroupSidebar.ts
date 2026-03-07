import { useMemo } from "react";
import { useGroups } from "@/hooks/useGroups";
import { useRecentGroupIds } from "@/hooks/useRecentGroups";

export function useGroupSidebar(currentGroupId?: string) {
  const groupsQuery = useGroups();
  const groups = groupsQuery.data?.items ?? [];
  const recentGroupIds = useRecentGroupIds();

  const navItems = useMemo(
    () => [
      { label: "全部分组", to: "/groups", active: !currentGroupId },
      ...groups.map((group) => ({
        label: group.name,
        to: `/groups/${group.id}`,
        active: group.id === currentGroupId,
      })),
    ],
    [currentGroupId, groups],
  );

  const recentItems = useMemo(
    () =>
      recentGroupIds
        .map((groupId) => groups.find((group) => group.id === groupId))
        .filter((group): group is (typeof groups)[number] => Boolean(group))
        .map((group) => ({
          label: group.name,
          to: `/groups/${group.id}`,
          meta: `${group.resource_count} 个资源`,
          active: group.id === currentGroupId,
        })),
    [currentGroupId, groups, recentGroupIds],
  );

  return { groups, navItems, recentItems, groupsQuery };
}
