<script setup>
/**
 * "Listening" widget on /stp: rotates through recent Spotify plays.
 *
 * Two timers, both owned by useRotation(): a 5s rotation and a 2-minute re-poll
 * of Spotify.
 */
import Header from "@/components/text/Header.vue";
import { computed, onMounted } from "vue";
import { useSongsStore } from "@/stores/songs";
import { useAuthStore } from "@/stores/auth";
import { useRotation } from "@/js/useRotation";

const ROTATE_MS = 5000;
const REFRESH_MS = 120000;

const songsStore = useSongsStore();

const { idx, next: nextSong } = useRotation(() => songsStore.songs, ROTATE_MS, {
  refresh: () => songsStore.fetchSongs(),
  refreshMs: REFRESH_MS,
});

const song = computed(() => songsStore.songs[idx.value]);

// Admin-only reconnect: when spotifyNeedsReauth comes back true the backend
// has no Spotify client (OAuth never completed or the token was revoked), so
// the widget offers the OAuth start endpoint instead of the template song.
// Non-admins get none of this.
const auth = useAuthStore();
const showReconnect = computed(
  () => auth.loggedIn && auth.user.admin && songsStore.needsReauth,
);

// GET /api/spotify/auth returns { url } — the Spotify authorisation URL with
// a one-shot state nonce. Navigating to it runs the OAuth dance; the callback
// installs the client server-side. Errors are swallowed: the widget simply
// stays as it was.
async function reconnect() {
  try {
    const res = await fetch("/api/spotify/auth");
    if (!res.ok) return;
    const { url } = await res.json();
    window.location.href = url;
  } catch (err) {
    console.error("Cannot start Spotify re-authentication", err);
  }
}

onMounted(() => {
  songsStore.fetchSongs();
});
</script>

<template>
  <div class="listening-wrapper">
    <div class="header-wrapper">
      <Header>Listening To</Header>
    </div>
    <div class="content-scroll">
      <div v-if="showReconnect" class="flex flex-col items-center">
        <p class="text-center">Spotify needs reconnecting.</p>
        <button
          type="button"
          class="underline text-mid cursor-pointer bg-transparent border-0"
          @click="reconnect"
        >
          Reconnect Spotify
        </button>
      </div>
      <Transition name="fade" v-else>
        <div
          @click="nextSong"
          :key="song.track.name"
          class="flex flex-col items-center"
        >
          <img
            :src="song.track.album.images[0].url"
            :alt="song.track.album.name + ' album art'"
          />
          <p class="text-center"><strong>Song:</strong> {{ song.track.name }}</p>
          <p class="text-center">
            <strong>Artist:</strong> {{ song.track.artists[0].name }}
          </p>
        </div>
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.listening-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.header-wrapper {
  flex-shrink: 0;
  display: flex;
  justify-content: center;
  padding-top: 0.5rem;
}

.content-scroll {
  position: relative;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

img {
  width: 70%;
  max-width: 100%;
  height: auto;
}
p {
  width: 100%;
  margin: 0 auto;
}
</style>
