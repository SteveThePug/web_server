/**
 * The single source of truth for the home page.
 *
 * One GraphQL query fetches everything the /stp widgets need (posts, favorites,
 * activities, Spotify, rowing, bookmarks, Gitea feed, Steam, current user) in a
 * single round trip, instead of each widget issuing its own request. The
 * per-domain stores (posts, favorites, activity, songs, steam, auth) are thin
 * views over this one — they `watch` their slice of it rather than fetching.
 *
 * fetchAll() is called once at the bottom of the setup function, i.e. the first
 * time any component calls useHomeDataStore(). That is why no component has to
 * kick off the initial load, and why `loaded` is the flag the router guard waits
 * on to know whether `me` has been resolved yet.
 */

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
  const steamStatus = ref(null);
  const bookmarks = ref([]);
  const radioLive = ref(false);

  /**
   * Fetch every home-page dataset in one GraphQL round trip, plus the radio
   * liveness probe in parallel. Sets `loaded` on success and `error` on
   * failure; never throws. Safe to call again to refresh.
   */
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
      loaded.value = true;
    } catch (err) {
      console.error("HomeData fetch failed:", err);
      error.value = err;
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
