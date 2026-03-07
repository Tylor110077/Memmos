import { useEffect, useState } from "react";

const STORAGE_KEY = "kg.recent-groups";
const RECENT_LIMIT = 5;
const RECENT_GROUPS_EVENT = "kg:recent-groups";
let memoryRecentGroupIds: string[] = [];

function isBrowser() {
  return typeof window !== "undefined";
}

function hasStorageApi() {
  return (
    isBrowser() &&
    typeof window.localStorage?.getItem === "function" &&
    typeof window.localStorage?.setItem === "function"
  );
}

export function readRecentGroupIds() {
  if (!isBrowser()) return memoryRecentGroupIds;

  try {
    if (!hasStorageApi()) {
      return memoryRecentGroupIds;
    }

    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const value = JSON.parse(raw);
    return Array.isArray(value) ? value.filter((item): item is string => typeof item === "string") : [];
  } catch {
    return memoryRecentGroupIds;
  }
}

function writeRecentGroupIds(groupIds: string[]) {
  memoryRecentGroupIds = groupIds;
  if (!isBrowser()) return;

  if (hasStorageApi()) {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(groupIds));
  }
  window.dispatchEvent(new CustomEvent(RECENT_GROUPS_EVENT));
}

export function markRecentGroup(groupId: string) {
  if (!groupId || !isBrowser()) return;

  const next = [groupId, ...readRecentGroupIds().filter((item) => item !== groupId)].slice(0, RECENT_LIMIT);
  writeRecentGroupIds(next);
}

export function useRecentGroupIds() {
  const [groupIds, setGroupIds] = useState<string[]>(() => readRecentGroupIds());

  useEffect(() => {
    function syncGroupIds() {
      setGroupIds(readRecentGroupIds());
    }

    window.addEventListener("storage", syncGroupIds);
    window.addEventListener(RECENT_GROUPS_EVENT, syncGroupIds);
    return () => {
      window.removeEventListener("storage", syncGroupIds);
      window.removeEventListener(RECENT_GROUPS_EVENT, syncGroupIds);
    };
  }, []);

  return groupIds;
}

export function useTrackRecentGroup(groupId?: string) {
  useEffect(() => {
    if (!groupId) return;
    markRecentGroup(groupId);
  }, [groupId]);
}

export function seedRecentGroups(groupIds: string[]) {
  writeRecentGroupIds(groupIds);
}

export function clearRecentGroups() {
  writeRecentGroupIds([]);
}
