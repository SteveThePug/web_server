/**
 * Recently played Spotify tracks for the Listening widget.
 *
 * Unlike the other derived stores this one also has its own fetchSongs(), because
 * Listening.vue re-polls Spotify every 2 minutes and re-running the whole
 * homeData query for that would be wasteful.
 */

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
  // Mirror the shared home query. `immediate` matters: homeData may already
  // have loaded by the time this store is first used, and a plain watch would
  // never fire for that existing value. The `length > 0` test keeps the
  // placeholder on screen when the backend returns nothing.
  watch(
    () => homeData.spotifyRecent,
    (newSongs) => {
      if (newSongs.length > 0) {
        songs.value = newSongs;
      }
    },
    { immediate: true },
  );

  /** Re-poll Spotify alone (cheaper than fetchAll). Keeps the old list on error. */
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
