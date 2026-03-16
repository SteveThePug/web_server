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
  watch(
    () => homeData.favorites,
    (newFavorites) => {
      if (newFavorites.length > 0) {
        favorites.value = newFavorites;
      }
    },
    { immediate: true },
  );

  async function fetchFavorites() {
    await homeData.fetchAll();
  }

  return {
    favorites,
    favoritesCount,
    fetchFavorites,
  };
});
