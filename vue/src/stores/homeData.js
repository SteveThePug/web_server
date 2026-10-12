/**
 * The single source of truth for the home page.
 *
 * One GraphQL query fetches everything the /stp widgets need (posts, favorites,
 * activities, Spotify, rowing, bookmarks, Gitea feed, Steam, current user) in a
 * single round trip, instead of each widget issuing its own request. The
 * per-domain stores (posts, favorites, activity, songs, steam) are thin views
 * over this one — they `watch` their slice of it rather than fetching. The
 * session user is the exception: `me` is pushed into stores/auth.js, which owns
 * it and can also fetch it alone (the router guard does not wait on this query).
 *
 * fetchAll() is called once at the bottom of the setup function, i.e. the first
 * time any component calls useHomeDataStore(). That is why no component has to
 * kick off the initial load.
 *
 * Two flags, because "finished" and "succeeded" are different questions:
 *   `settled` — a fetch has completed, either way. Wait on / show a spinner
 *               against THIS; it always becomes true.
 *   `loaded`  — a fetch has succeeded at least once, so the data refs are real.
 * Anything that waited on `loaded` alone would wait forever when the query
 * fails outright, which is how the router guard used to hang.
 */

import { defineStore } from "pinia";
import { ref } from "vue";
import { gql } from "@/graphql";
import axios from "axios";
import { useAuthStore } from "@/stores/auth";

export const useHomeDataStore = defineStore("homeData", () => {
  const loaded = ref(false);
  const settled = ref(false);
  const error = ref(null);

  const me = ref(null);
  const posts = ref([]);
  const favorites = ref([]);
  const activities = ref([]);
  const spotifyRecent = ref([]);
  const rowingSessions = ref([]);
  const gitFeed = ref(null);
  const steamStatus = ref(null);
  const bookmarks = ref([]);
  const radioLive = ref(false);

  /**
   * Fetch every home-page dataset in one GraphQL round trip, plus the radio
   * liveness probe in parallel. Sets `loaded` on success and `error` on
   * failure, and `settled` either way; never throws. Safe to call again to
   * refresh.
   */
  async function fetchAll() {
    const auth = useAuthStore();
    // Read before the request goes out: if the user logs in or out while it is
    // in flight, the `me` it returns is stale and auth will ignore it.
    const sessionAt = auth.sessionEpoch();
    try {
      const [data] = await Promise.all([
        gql(`
          query HomeData {
            posts { id title content createdAt updatedAt author { id username } }
            favorites { id type name link createdAt }
            activities { id type name link createdAt }
            spotifyRecent { track { name album { name images { url } } artists { name } } playedAt }
            rowingSessions { id date time distance timePer500m calories }
            bookmarks { id category name link }
            giteaFeed { avatarUrl repoUrl repoName opType commitMessage createdAt }
            steamStatus { online recentGames { appId name playtime2Weeks playtimeForever headerImageUrl } }
            me { id username admin }
          }
        `),
        fetchRadioStatus(),
      ]);
      posts.value = data.posts;
      favorites.value = data.favorites;
      activities.value = data.activities;
      spotifyRecent.value = data.spotifyRecent || [];
      bookmarks.value = data.bookmarks || [];
      rowingSessions.value = data.rowingSessions;
      gitFeed.value = data.giteaFeed || null;
      steamStatus.value = data.steamStatus || null;
      me.value = data.me || null;
      // The home query already says who is signed in, so hand it to the auth
      // store rather than have it issue its own `me` query on this page.
      auth.adoptSession(me.value, sessionAt);
      loaded.value = true;
      // A refresh that works clears the failure a previous one left behind.
      error.value = null;
    } catch (err) {
      console.error("HomeData fetch failed:", err);
      error.value = err;
    } finally {
      settled.value = true;
    }
  }

  /**
   * Probe the Icecast stream with a HEAD request and set `radioLive`.
   * A failure (404/connection refused) means "offline", not an error.
   */
  async function fetchRadioStatus() {
    try {
      await axios.head("/radio/stream");
      radioLive.value = true;
    } catch {
      radioLive.value = false;
    }
  }

  // Kick off the initial load as a side effect of the first useHomeDataStore()
  // call. Pinia setup stores run once, so this fires exactly once per page load.
  fetchAll();

  return {
    loaded,
    settled,
    error,
    me,
    bookmarks,
    posts,
    favorites,
    activities,
    spotifyRecent,
    rowingSessions,
    gitFeed,
    steamStatus,
    radioLive,
    fetchAll,
    fetchRadioStatus,
  };
});
