export type GroupEventName =
  | "resource.status.changed"
  | "resource.completed"
  | "resource.failed"
  | "framework_graph.updated"
  | "graph.node.expanded";

type EventSourceLike = Pick<EventSource, "addEventListener" | "close"> & {
  onerror: ((this: EventSource, ev: Event) => unknown) | null;
  onopen: ((this: EventSource, ev: Event) => unknown) | null;
};

type SubscribeOptions = {
  EventSourceImpl?: new (url: string) => EventSourceLike;
  maxRetries?: number;
  reconnectDelayMs?: number;
};

const eventNames: GroupEventName[] = [
  "resource.status.changed",
  "resource.completed",
  "resource.failed",
  "framework_graph.updated",
  "graph.node.expanded",
];

export function subscribeGroupEvents(
  groupId: string,
  onEvent: (event: GroupEventName, payload: unknown) => void,
  options: SubscribeOptions = {},
) {
  const EventSourceImpl = options.EventSourceImpl ?? EventSource;
  const maxRetries = options.maxRetries ?? 3;
  const reconnectDelayMs = options.reconnectDelayMs ?? 1000;

  let source: EventSourceLike | null = null;
  let destroyed = false;
  let retries = 0;
  let reconnectTimer: number | undefined;

  const connect = () => {
    if (destroyed) return;

    source = new EventSourceImpl(`/api/v1/events/groups/${groupId}`);

    source.onopen = () => {
      retries = 0;
    };

    eventNames.forEach((eventName) => {
      source?.addEventListener(eventName, (event) => {
        const payload = JSON.parse((event as MessageEvent).data);
        onEvent(eventName, payload);
      });
    });

    source.onerror = () => {
      source?.close();
      if (destroyed || retries >= maxRetries) {
        return;
      }

      retries += 1;
      reconnectTimer = window.setTimeout(connect, reconnectDelayMs * retries);
    };
  };

  connect();

  return () => {
    destroyed = true;
    if (reconnectTimer) {
      window.clearTimeout(reconnectTimer);
    }
    source?.close();
  };
}
