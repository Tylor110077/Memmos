import { create } from "zustand";

type GraphWorkbenchState = {
  selectedNodeId?: string;
  includeExpansion: boolean;
  maxLevel?: number;
  setSelectedNodeId: (nodeId?: string) => void;
  setIncludeExpansion: (value: boolean) => void;
  setMaxLevel: (value?: number) => void;
  reset: () => void;
};

export const useGraphWorkbenchStore = create<GraphWorkbenchState>((set) => ({
  selectedNodeId: undefined,
  includeExpansion: true,
  maxLevel: 3,
  setSelectedNodeId: (selectedNodeId) => set({ selectedNodeId }),
  setIncludeExpansion: (includeExpansion) => set({ includeExpansion }),
  setMaxLevel: (maxLevel) => set({ maxLevel }),
  reset: () => set({ selectedNodeId: undefined, includeExpansion: true, maxLevel: 3 }),
}));
