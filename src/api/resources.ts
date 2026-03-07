import { apiRequest } from "@/api/client";
import type { Paginated, ResourceDetail, ResourceMutationResult, ResourceSummary } from "@/api/types";
import { normalizeResourceDetail, normalizeResourceSummaries } from "@/api/liveAdapters";

export async function getResources(groupId: string) {
  const response = await apiRequest<Paginated<ResourceSummary> | ResourceSummary[]>(`/groups/${groupId}/resources`);
  return normalizeResourceSummaries(response);
}

export async function getResourceDetail(resourceId: string, groupId?: string) {
  const path =
    import.meta.env.VITE_API_MODE !== "mock" && groupId
      ? `/groups/${groupId}/resources/${resourceId}`
      : `/resources/${resourceId}`;
  const response = await apiRequest<ResourceDetail>(path);
  return normalizeResourceDetail(response);
}

export function uploadResource(groupId: string, file: File, name?: string) {
  const formData = new FormData();
  formData.append("file", file);
  if (name) {
    formData.append("name", name);
  }

  const path =
    import.meta.env.VITE_API_MODE !== "mock" ? `/groups/${groupId}/resources/upload` : `/groups/${groupId}/resources`;
  return apiRequest<ResourceMutationResult>(path, {
    method: "POST",
    body: formData,
  });
}

export function createWebResource(groupId: string, payload: { url: string; name?: string }) {
  return apiRequest<ResourceMutationResult>(`/groups/${groupId}/web-resources`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function retryResource(resourceId: string) {
  return apiRequest<ResourceMutationResult>(`/resources/${resourceId}/retry`, {
    method: "POST",
    body: JSON.stringify({}),
  });
}

export function deleteResource(resourceId: string) {
  return apiRequest<{ success: boolean }>(`/resources/${resourceId}`, {
    method: "DELETE",
  });
}
