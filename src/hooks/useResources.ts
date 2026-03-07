import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createWebResource,
  deleteResource,
  getResourceDetail,
  getResources,
  retryResource,
  uploadResource,
} from "@/api/resources";
import { queryKeys } from "@/api/queryKeys";

export function useResources(groupId: string) {
  return useQuery({
    queryKey: queryKeys.resources(groupId),
    queryFn: () => getResources(groupId),
    enabled: Boolean(groupId),
  });
}

export function useResourceDetail(resourceId: string, groupId?: string) {
  return useQuery({
    queryKey: [...queryKeys.resourceDetail(resourceId), groupId ?? ""] as const,
    queryFn: () => getResourceDetail(resourceId, groupId),
    enabled: Boolean(resourceId),
  });
}

export function useUploadResource(groupId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: { file: File; name?: string }) => uploadResource(groupId, payload.file, payload.name),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.resources(groupId) });
      void queryClient.invalidateQueries({ queryKey: queryKeys.groupDetail(groupId) });
    },
  });
}

export function useCreateWebResource(groupId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: { url: string; name?: string }) => createWebResource(groupId, payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.resources(groupId) });
      void queryClient.invalidateQueries({ queryKey: queryKeys.groupDetail(groupId) });
    },
  });
}

export function useRetryResource(groupId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: retryResource,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.resources(groupId) });
    },
  });
}

export function useDeleteResource(groupId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteResource,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.resources(groupId) });
      void queryClient.invalidateQueries({ queryKey: queryKeys.groupDetail(groupId) });
    },
  });
}
