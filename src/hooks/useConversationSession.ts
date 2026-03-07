const STORAGE_KEY = "goaipj-conversation-sessions";

type ConversationSessionMap = Record<string, string>;

function getStorage() {
  if (typeof window === "undefined") {
    return null;
  }

  const storage = window.localStorage;
  if (!storage || typeof storage.getItem !== "function" || typeof storage.setItem !== "function") {
    return null;
  }

  return storage;
}

function getSessionKey(groupId: string, resourceId: string) {
  return `${groupId}:${resourceId}`;
}

function readSessionMap(): ConversationSessionMap {
  const raw = getStorage()?.getItem(STORAGE_KEY);
  if (!raw) {
    return {};
  }

  try {
    const parsed = JSON.parse(raw) as ConversationSessionMap;
    return parsed ?? {};
  } catch {
    return {};
  }
}

function writeSessionMap(map: ConversationSessionMap) {
  getStorage()?.setItem(STORAGE_KEY, JSON.stringify(map));
}

export function readConversationSession(groupId: string, resourceId: string) {
  return readSessionMap()[getSessionKey(groupId, resourceId)] ?? "";
}

export function persistConversationSession(groupId: string, resourceId: string, conversationId: string) {
  const current = readSessionMap();
  current[getSessionKey(groupId, resourceId)] = conversationId;
  writeSessionMap(current);
}

export function clearConversationSession(groupId: string, resourceId: string) {
  const current = readSessionMap();
  delete current[getSessionKey(groupId, resourceId)];
  writeSessionMap(current);
}
