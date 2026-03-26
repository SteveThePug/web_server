import { defineStore } from "pinia";
import { ref, watch } from "vue";
import { gql } from "@/graphql";
import { useHomeDataStore } from "@/stores/homeData";

export const useSteamStore = defineStore("steam", () => {
  const steamStatus = ref({ online: false, recentGames: [] });

  const homeData = useHomeDataStore();
  watch(
    () => homeData.steamStatus,
    (newStatus) => {
      if (newStatus) {
        steamStatus.value = newStatus;
      }
    },
    { immediate: true },
  );

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
