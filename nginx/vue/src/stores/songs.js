import { defineStore } from "pinia";
import { ref, computed } from "vue";
import axios from "axios";

const song_template = {
  id: 1,
  track: {
    id: 1,
    name: "^_^",
    album: { images: [{ url: "/img/Untitled.png" }] },
    artists: [{ name: ">_<" }],
  },
};

export const useSongsStore = defineStore("songs", () => {
  const songs = ref([song_template]);

  const songsCount = computed(() => songs.value.length);

  fetchSongs();

  let refreshId = null;
  refreshId = setTimeout(fetchSongs, 120000);

  async function fetchSongs() {
    try {
      const res = await axios.get("/api/spotify/recent");
      if (!Array.isArray(res.data)) {
        throw new Error("Invalid response from Spotify API");
      }
      songs.value = res.data;
      refreshId = setTimeout(fetchSongs, 120000);
      console.log("lol");
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
