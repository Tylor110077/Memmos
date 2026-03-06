import { apiRequest } from "@/api/client";
import type { Group, GroupDetail, Paginated } from "@/api/types";

export function getGroups(keyword = "") {
  const query = new URLSearchParams({ page: "1", page_size: "20" });
  if (keyword) query.set("keyword", keyword);
  return apiRequest<Paginated<Group>>(`/groups?${query.toString()}`);
}

export function getGroupDetail(groupId: string) {
  return apiRequest<GroupDetail>(`/groups/${groupId}`);
}

export function createGroup(payload: { name: string; description?: string }) {
  return apiRequest<Group>("/groups", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function updateGroup(groupId: string, payload: { name: string; description?: string }) {
  return apiRequest<Group>(`/groups/${groupId}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export function deleteGroup(groupId: string) {
  return apiRequest<{ success: boolean }>(`/groups/${groupId}`, {
    method: "DELETE",
  });
}
