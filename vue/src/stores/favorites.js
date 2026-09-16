/**
 * Favourites for the Favorites widget. A view over homeData.favorites; see
 * stores/posts.js for the placeholder-then-overwrite pattern.
 */

import { defineStore } from "pinia";
import { computed, ref, watch } from "vue";
import { useHomeDataStore } from "@/stores/homeData";

const favorite_template = {
  type: "favorite",
  name: "nameof",
  createdAt: Date.now(),
};

export const useFavoritesStore = defineStore("favorites", () => {
  const favorites = ref([favorite_template]);

  const favoritesCount = computed(() => favorites.value.length);

  const homeData = useHomeDataStore();
  // Mirror the shared home query. `immediate` matters: homeData may already
  // have loaded by the time this store is first used, and a plain watch would
  // never fire for that existing value. The `length > 0` test keeps the
  // placeholder on screen when the backend returns nothing.
  watch(
    () => homeData.favorites,
    (newFavorites) => {
      if (newFavorites.length > 0) {
        favorites.value = newFavorites;
      }
    },
    { immediate: true },
  );

  /** Refresh by re-running the whole home query; the watch above applies it. */
  async function fetchFavorites() {
    await homeData.fetchAll();
  }

  return {
    favorites,
    favoritesCount,
    fetchFavorites,
  };
});
