import { defineStore } from "pinia";
import { ref, computed, watch } from "vue";
import { gql } from "@/graphql";
import { useHomeDataStore } from "@/stores/homeData";

const song_template = {
  track: {
    name: "^_^",
    album: { name: "", images: [{ url: "/img/Untitled.png" }] },
    artists: [{ name: ">_<" }],
  },
};

export const useSongsStore = defineStore("songs", () => {
  const songs = ref([song_template]);

  const songsCount = computed(() => songs.value.length);

  const homeData = useHomeDataStore();
  watch(
    () => homeData.spotifyRecent,
    (newSongs) => {
      if (newSongs.length > 0) {
        songs.value = newSongs;
      }
    },
    { immediate: true },
  );

  async function fetchSongs() {
    try {
      const data = await gql(`
        query {
          spotifyRecent {
            track {
              name
              album { name images { url } }
              artists { name }
            }
            playedAt
          }
        }
      `);
      if (Array.isArray(data.spotifyRecent) && data.spotifyRecent.length > 0) {
        songs.value = data.spotifyRecent;
      }
    } catch (err) {
      console.error("Cannot connect to Spotify API", err);
    }
  }

  return {
    songs,

    songsCount,

    fetchSongs,
  };
});
