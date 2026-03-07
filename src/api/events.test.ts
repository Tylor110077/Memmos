import { afterEach, describe, expect, it, vi } from "vitest";
import { subscribeGroupEvents } from "@/api/events";

class MockEventSource {
  static instances: MockEventSource[] = [];
  static reset() {
    MockEventSource.instances = [];
  }

  onerror: ((this: EventSource, ev: Event) => unknown) | null = null;
  onopen: ((this: EventSource, ev: Event) => unknown) | null = null;
  listeners = new Map<string, Array<(event: MessageEvent) => void>>();
  closed = false;

  constructor(public url: string) {
    MockEventSource.instances.push(this);
  }

  addEventListener(name: string, handler: (event: MessageEvent) => void) {
    const current = this.listeners.get(name) ?? [];
    current.push(handler);
    this.listeners.set(name, current);
  }

  emit(name: string, payload: unknown) {
    this.listeners.get(name)?.forEach((handler) =>
      handler({ data: JSON.stringify(payload) } as MessageEvent),
    );
  }

  close() {
    this.closed = true;
  }
}

describe("subscribeGroupEvents", () => {
  afterEach(() => {
    MockEventSource.reset();
    vi.useRealTimers();
  });

  it("forwards typed group events", () => {
    const handler = vi.fn();

    const dispose = subscribeGroupEvents("grp_agent", handler, {
      EventSourceImpl: MockEventSource as any,
    });

    const source = MockEventSource.instances[0];
    source.emit("resource.completed", { group_id: "grp_agent", resource_id: "res_1", graph_id: "graph_1", status: "completed" });

    expect(handler).toHaveBeenCalledWith("resource.completed", {
      group_id: "grp_agent",
      resource_id: "res_1",
      graph_id: "graph_1",
      status: "completed",
    });

    dispose();
    expect(source.closed).toBe(true);
  });

  it("reconnects after an error until disposed", () => {
    vi.useFakeTimers();
    const handler = vi.fn();

    const dispose = subscribeGroupEvents("grp_agent", handler, {
      EventSourceImpl: MockEventSource as any,
      reconnectDelayMs: 10,
      maxRetries: 2,
    });

    expect(MockEventSource.instances).toHaveLength(1);
    MockEventSource.instances[0].onerror?.call({} as EventSource, new Event("error"));
    vi.advanceTimersByTime(10);

    expect(MockEventSource.instances).toHaveLength(2);

    dispose();
    MockEventSource.instances[1].onerror?.call({} as EventSource, new Event("error"));
    vi.advanceTimersByTime(10);

    expect(MockEventSource.instances).toHaveLength(2);
  });
});
