import { create } from "zustand";

import type { Item } from "../api/types";

export type SearchScope =
  | { type: "feed"; feedId: string }
  | { type: "all" };

type UiState = {
  drawerOpen: boolean;
  searchOpen: boolean;
  searchQuery: string;
  searchScope: SearchScope;
  readerItem: Item | null;
  setDrawerOpen: (open: boolean) => void;
  toggleDrawer: () => void;
  setSearchOpen: (open: boolean) => void;
  setSearchQuery: (query: string) => void;
  setSearchScope: (scope: SearchScope) => void;
  openReader: (item: Item) => void;
  closeReader: () => void;
  closeSearch: () => void;
};

export const useUiStore = create<UiState>((set) => ({
  drawerOpen: false,
  searchOpen: false,
  searchQuery: "",
  searchScope: { type: "all" },
  readerItem: null,
  setDrawerOpen: (drawerOpen) => set({ drawerOpen }),
  toggleDrawer: () => set((state) => ({ drawerOpen: !state.drawerOpen })),
  setSearchOpen: (searchOpen) => set({ searchOpen }),
  setSearchQuery: (searchQuery) => set({ searchQuery }),
  setSearchScope: (searchScope) => set({ searchScope }),
  openReader: (readerItem) => set({ readerItem }),
  closeReader: () => set({ readerItem: null }),
  closeSearch: () => set({ searchOpen: false, searchQuery: "" })
}));
