import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it } from "vitest";
import { queryKeys } from "@/api/queryKeys";
import { applyGroupEventToCache } from "@/hooks/useGroupEvents";
import type { ResourceDetail } from "@/api/types";

describe("applyGroupEventToCache", () => {
  it("updates resource status caches on status change", () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(queryKeys.resources("grp_agent"), {
      items: [
        {
          id: "res_1",
          group_id: "grp_agent",
          name: "demo.pdf",
          resource_type: "pdf",
          status: "uploaded",
          error_message: null,
          created_at: "2026-03-06T12:00:00Z",
          updated_at: "2026-03-06T12:00:00Z",
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    });
    queryClient.setQueryData<ResourceDetail>(queryKeys.resourceDetail("res_1"), {
      id: "res_1",
      group_id: "grp_agent",
      name: "demo.pdf",
      resource_type: "pdf",
      source_uri: null,
      status: "uploaded",
      error_code: null,
      error_message: null,
      failed_stage: null,
      latest_job_id: "job_1",
      latest_graph_id: null,
      created_at: "2026-03-06T12:00:00Z",
      updated_at: "2026-03-06T12:00:00Z",
    });

    applyGroupEventToCache(queryClient as any, "resource.status.changed", {
      group_id: "grp_agent",
      resource_id: "res_1",
      status: "graph_generating",
      updated_at: "2026-03-06T12:05:00Z",
    });

    expect((queryClient.getQueryData(queryKeys.resources("grp_agent")) as any).items[0].status).toBe("graph_generating");
    expect((queryClient.getQueryData(queryKeys.resourceDetail("res_1")) as ResourceDetail).status).toBe("graph_generating");
  });
});
