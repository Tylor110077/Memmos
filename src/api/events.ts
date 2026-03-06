export function subscribeGroupEvents(groupId: string, onEvent: (event: string, payload: unknown) => void) {
  const source = new EventSource(`/api/v1/events/groups/${groupId}`);

  const eventNames = [
    "resource.status.changed",
    "resource.completed",
    "resource.failed",
    "framework_graph.updated",
    "graph.node.expanded",
  ];

  eventNames.forEach((eventName) => {
    source.addEventListener(eventName, (event) => {
      onEvent(eventName, JSON.parse((event as MessageEvent).data));
    });
  });

  return () => source.close();
}
