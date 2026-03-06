import { apiRequest } from "@/api/client";
import type { Paginated, ResourceDetail, ResourceMutationResult, ResourceSummary } from "@/api/types";

export function getResources(groupId: string) {
  return apiRequest<Paginated<ResourceSummary>>(`/groups/${groupId}/resources`);
}

export function getResourceDetail(resourceId: string) {
  return apiRequest<ResourceDetail>(`/resources/${resourceId}`);
}

export function uploadResource(groupId: string, file: File, name?: string) {
  const formData = new FormData();
  formData.append("file", file);
  if (name) {
    formData.append("name", name);
  }

  return apiRequest<ResourceMutationResult>(`/groups/${groupId}/resources`, {
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
