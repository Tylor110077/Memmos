import { create } from "zustand";
import type { NodeType } from "@/api/types";

type GraphWorkbenchState = {
  selectedNodeId?: string;
  includeExpansion: boolean;
  maxLevel?: number;
  searchTerm: string;
  nodeTypeFilter: NodeType[];
  setSelectedNodeId: (nodeId?: string) => void;
  setIncludeExpansion: (value: boolean) => void;
  setMaxLevel: (value?: number) => void;
  setSearchTerm: (value: string) => void;
  toggleNodeType: (value: NodeType) => void;
  clearNodeTypeFilter: () => void;
  reset: () => void;
};

export const useGraphWorkbenchStore = create<GraphWorkbenchState>((set) => ({
  selectedNodeId: undefined,
  includeExpansion: true,
  maxLevel: 3,
  searchTerm: "",
  nodeTypeFilter: [],
  setSelectedNodeId: (selectedNodeId) => set({ selectedNodeId }),
  setIncludeExpansion: (includeExpansion) => set({ includeExpansion }),
  setMaxLevel: (maxLevel) => set({ maxLevel }),
  setSearchTerm: (searchTerm) => set({ searchTerm }),
  toggleNodeType: (value) =>
    set((state) => ({
      nodeTypeFilter: state.nodeTypeFilter.includes(value)
        ? state.nodeTypeFilter.filter((item) => item !== value)
        : [...state.nodeTypeFilter, value],
    })),
  clearNodeTypeFilter: () => set({ nodeTypeFilter: [] }),
  reset: () => set({ selectedNodeId: undefined, includeExpansion: true, maxLevel: 3, searchTerm: "", nodeTypeFilter: [] }),
}));
