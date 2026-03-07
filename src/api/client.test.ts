import { describe, expect, it } from "vitest";
import { apiRequest } from "@/api/client";

describe("apiRequest", () => {
  it("parses a successful response payload", async () => {
    const response = await apiRequest<{ items: Array<{ id: string }> }>("/groups?page=1&page_size=20");
    expect(response.items[0]?.id).toBe("grp_agent");
  });

  it("throws on failed responses", async () => {
    await expect(apiRequest("/groups/unknown")).rejects.toThrow("group not found");
  });
});
