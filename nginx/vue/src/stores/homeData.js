import { defineStore } from "pinia";
import { ref } from "vue";
import { gql } from "@/graphql";

export const useHomeDataStore = defineStore("homeData", () => {
  const loaded = ref(false);
  const error = ref(null);

  const me = ref(null);
  const posts = ref([]);
  const favorites = ref([]);
  const activities = ref([]);
  const spotifyRecent = ref([]);

  async function fetchAll() {
    try {
      const data = await gql(`
        query HomeData {
          posts { id title content createdAt updatedAt author { id username } }
          favorites { id type name link createdAt }
          activities { id type name link createdAt }
          spotifyRecent { track { name album { name images { url } } artists { name } } playedAt }
          me { id username admin }
        }
      `);
      posts.value = data.posts;
      favorites.value = data.favorites;
      activities.value = data.activities;
      spotifyRecent.value = data.spotifyRecent;
      me.value = data.me || null;
      loaded.value = true;
    } catch (err) {
      console.error("HomeData fetch failed:", err);
      error.value = err;
    }
  }

  fetchAll();

  return {
    loaded,
    error,
    me,
    posts,
    favorites,
    activities,
    spotifyRecent,
    fetchAll,
  };
});
