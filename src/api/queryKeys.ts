export const queryKeys = {
  groups: (keyword = "") => ["groups", keyword] as const,
  groupDetail: (groupId: string) => ["groups", groupId] as const,
  resources: (groupId: string) => ["groups", groupId, "resources"] as const,
  resourceDetail: (resourceId: string) => ["resources", resourceId] as const,
  resourceGraph: (resourceId: string, params: { includeExpansion: boolean; maxLevel?: number }) =>
    ["resources", resourceId, "graph", params] as const,
  frameworkGraph: (groupId: string, maxLevel = 2) => ["groups", groupId, "framework-graph", maxLevel] as const,
  nodeDetail: (graphId: string, nodeId: string) => ["graphs", graphId, "nodes", nodeId] as const,
  conversation: (conversationId: string) => ["conversations", conversationId] as const,
  job: (jobId: string) => ["jobs", jobId] as const,
};
