/**
 * Favourites for the Favorites widget. A view over homeData.favorites, built by
 * createMirrorStore(); see stores/createMirrorStore.js for the
 * placeholder-then-overwrite pattern.
 */

import { defineStore } from "pinia";
import { createMirrorStore } from "@/stores/createMirrorStore";

const favorite_template = {
  type: "favorite",
  name: "nameof",
  createdAt: Date.now(),
};

export const useFavoritesStore = defineStore("favorites", () => {
  const { items, count, fetch } = createMirrorStore({
    slice: "favorites",
    template: favorite_template,
  });

  return {
    favorites: items,
    favoritesCount: count,
    fetchFavorites: fetch,
  };
});
