import { create } from "zustand";

export type SearchScope =
  | { type: "feed"; feedId: string }
  | { type: "all" };

type UiState = {
  drawerOpen: boolean;
  searchOpen: boolean;
  searchQuery: string;
  searchScope: SearchScope;
  setDrawerOpen: (open: boolean) => void;
  toggleDrawer: () => void;
  setSearchOpen: (open: boolean) => void;
  setSearchQuery: (query: string) => void;
  setSearchScope: (scope: SearchScope) => void;
  closeSearch: () => void;
};

export const useUiStore = create<UiState>((set) => ({
  drawerOpen: false,
  searchOpen: false,
  searchQuery: "",
  searchScope: { type: "all" },
  setDrawerOpen: (drawerOpen) => set({ drawerOpen }),
  toggleDrawer: () => set((state) => ({ drawerOpen: !state.drawerOpen })),
  setSearchOpen: (searchOpen) => set({ searchOpen }),
  setSearchQuery: (searchQuery) => set({ searchQuery }),
  setSearchScope: (searchScope) => set({ searchScope }),
  closeSearch: () => set({ searchOpen: false, searchQuery: "" })
}));
