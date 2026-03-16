import { defineStore } from "pinia";
import { ref } from "vue";
import { gql } from "@/graphql";
import axios from "axios";

export const useHomeDataStore = defineStore("homeData", () => {
  const loaded = ref(false);
  const error = ref(null);

  const me = ref(null);
  const posts = ref([]);
  const favorites = ref([]);
  const activities = ref([]);
  const spotifyRecent = ref([]);
  const rowingSessions = ref([]);
  const gitFeed = ref(null);
  const radioLive = ref(false);

  async function fetchAll() {
    try {
      const [data] = await Promise.all([
        gql(`
          query HomeData {
            posts { id title content createdAt updatedAt author { id username } }
            favorites { id type name link createdAt }
            activities { id type name link createdAt }
            spotifyRecent { track { name album { name images { url } } artists { name } } playedAt }
            rowingSessions { id date time distance timePer500m calories }
            me { id username admin }
          }
        `),
        fetchGitFeed(),
        fetchRadioStatus(),
      ]);
      posts.value = data.posts;
      favorites.value = data.favorites;
      activities.value = data.activities;
      spotifyRecent.value = data.spotifyRecent;
      rowingSessions.value = data.rowingSessions;
      me.value = data.me || null;
      loaded.value = true;
    } catch (err) {
      console.error("HomeData fetch failed:", err);
      error.value = err;
    }
  }

  async function fetchGitFeed() {
    try {
      const res = await axios.get("/gitea/api/v1/users/adamf/activities/feeds?limit=1");
      gitFeed.value = res.data[0] || null;
    } catch {
      gitFeed.value = null;
    }
  }

  async function fetchRadioStatus() {
    try {
      await axios.head("/radio/stream");
      radioLive.value = true;
    } catch {
      radioLive.value = false;
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
    rowingSessions,
    gitFeed,
    radioLive,
    fetchAll,
    fetchRadioStatus,
  };
});
