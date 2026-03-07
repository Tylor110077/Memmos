import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createGroup, deleteGroup, getGroupDetail, getGroups, updateGroup } from "@/api/groups";
import { queryKeys } from "@/api/queryKeys";

export function useGroups(keyword = "") {
  return useQuery({
    queryKey: queryKeys.groups(keyword),
    queryFn: () => getGroups(keyword),
  });
}

export function useCreateGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createGroup,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["groups"] });
    },
  });
}

export function useUpdateGroup(groupId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: { name: string; description?: string }) => updateGroup(groupId, payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["groups"] });
      void queryClient.invalidateQueries({ queryKey: queryKeys.groupDetail(groupId) });
    },
  });
}

export function useDeleteGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteGroup,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["groups"] });
    },
  });
}

export function useGroupDetail(groupId: string) {
  return useQuery({
    queryKey: queryKeys.groupDetail(groupId),
    queryFn: () => getGroupDetail(groupId),
    enabled: Boolean(groupId),
  });
}
