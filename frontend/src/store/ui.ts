import { create } from "zustand";

type UiState = {
  drawerOpen: boolean;
  searchOpen: boolean;
  searchQuery: string;
  setDrawerOpen: (open: boolean) => void;
  toggleDrawer: () => void;
  setSearchOpen: (open: boolean) => void;
  setSearchQuery: (query: string) => void;
  closeSearch: () => void;
};

export const useUiStore = create<UiState>((set) => ({
  drawerOpen: false,
  searchOpen: false,
  searchQuery: "",
  setDrawerOpen: (drawerOpen) => set({ drawerOpen }),
  toggleDrawer: () => set((state) => ({ drawerOpen: !state.drawerOpen })),
  setSearchOpen: (searchOpen) => set({ searchOpen }),
  setSearchQuery: (searchQuery) => set({ searchQuery }),
  closeSearch: () => set({ searchOpen: false, searchQuery: "" })
}));
