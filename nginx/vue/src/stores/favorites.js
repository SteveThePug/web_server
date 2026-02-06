import { defineStore } from "pinia";
import { computed, ref } from "vue";
import axios from "axios";

const favorite_template = {
  type: "favorite",
  name: "nameof",
  createdAt: Date.now(),
};

export const useFavoritesStore = defineStore("favorites", () => {
  const favorites = ref([favorite_template]);

  const favoritesCount = computed(() => favorites.value.length);

  async function fetchFavorites() {
    try {
      const res = await axios.get("/api/favorites");
      if (!Array.isArray(res.data)) {
        throw new Error("Invalid response from favorites API");
      }
      favorites.value = res.data;
    } catch (err) {
      console.error("Cannot connect to favorites API", err);
    }
  }

  fetchFavorites();

  return {
    favorites,
    favoritesCount,
    fetchFavorites,
  };
});
