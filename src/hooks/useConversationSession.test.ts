import { afterEach, beforeEach, describe, expect, it } from "vitest";
import {
  clearConversationSession,
  persistConversationSession,
  readConversationSession,
} from "@/hooks/useConversationSession";

describe("conversation session storage", () => {
  const storage = new Map<string, string>();

  beforeEach(() => {
    Object.defineProperty(window, "localStorage", {
      configurable: true,
      value: {
        getItem: (key: string) => storage.get(key) ?? null,
        setItem: (key: string, value: string) => {
          storage.set(key, value);
        },
      },
    });
  });

  afterEach(() => {
    storage.clear();
  });

  it("persists and clears conversation ids by group-resource pair", () => {
    persistConversationSession("grp_1", "res_1", "conv_1");
    persistConversationSession("grp_1", "res_2", "conv_2");

    expect(readConversationSession("grp_1", "res_1")).toBe("conv_1");
    expect(readConversationSession("grp_1", "res_2")).toBe("conv_2");

    clearConversationSession("grp_1", "res_1");

    expect(readConversationSession("grp_1", "res_1")).toBe("");
    expect(readConversationSession("grp_1", "res_2")).toBe("conv_2");
  });
});
