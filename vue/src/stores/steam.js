/**
 * Steam online status and recent games for the Steam widget. Mirrors
 * homeData.steamStatus, plus a standalone fetchSteam() used for the widget's own
 * 5-minute refresh.
 */

import { defineStore } from "pinia";
import { ref, watch } from "vue";
import { gql } from "@/graphql";
import { useHomeDataStore } from "@/stores/homeData";

export const useSteamStore = defineStore("steam", () => {
  const steamStatus = ref({ online: false, recentGames: [] });

  const homeData = useHomeDataStore();
  // Mirror the shared home query; `immediate` so a store created after
  // homeData has already loaded still picks up the value.
  watch(
    () => homeData.steamStatus,
    (newStatus) => {
      if (newStatus) {
        steamStatus.value = newStatus;
      }
    },
    { immediate: true },
  );

  /** Re-poll Steam alone (cheaper than fetchAll). Keeps the old value on error. */
  async function fetchSteam() {
    try {
      const data = await gql(`
        query {
          steamStatus {
            online
            recentGames { appId name playtime2Weeks playtimeForever headerImageUrl }
          }
        }
      `);
      if (data.steamStatus) {
        steamStatus.value = data.steamStatus;
      }
    } catch (err) {
      console.error("Failed to fetch Steam status", err);
    }
  }

  return {
    steamStatus,
    fetchSteam,
  };
});
