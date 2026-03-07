import { apiRequest } from "@/api/client";
import type { Group, GroupDetail, Paginated } from "@/api/types";
import { normalizeGroup, normalizeGroupDetail, normalizeGroups } from "@/api/liveAdapters";

export async function getGroups(keyword = "") {
  const query = new URLSearchParams({ page: "1", page_size: "20" });
  if (keyword) query.set("keyword", keyword);
  const response = await apiRequest<Paginated<Group> | Group[]>(`/groups?${query.toString()}`);
  return normalizeGroups(response);
}

export async function getGroupDetail(groupId: string) {
  const response = await apiRequest<GroupDetail>(`/groups/${groupId}`);
  return normalizeGroupDetail(response);
}

export async function createGroup(payload: { name: string; description?: string }) {
  const response = await apiRequest<Group>("/groups", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  return normalizeGroup(response);
}

export async function updateGroup(groupId: string, payload: { name: string; description?: string }) {
  const response = await apiRequest<Group>(`/groups/${groupId}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
  return normalizeGroup(response);
}

export function deleteGroup(groupId: string) {
  return apiRequest<{ success: boolean }>(`/groups/${groupId}`, {
    method: "DELETE",
  });
}
