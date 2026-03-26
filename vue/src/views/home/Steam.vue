<script setup>
import { onMounted, onUnmounted } from "vue";
import { storeToRefs } from "pinia";
import { useSteamStore } from "@/stores/steam";
import { useHomeDataStore } from "@/stores/homeData";
import Header from "@/components/text/Header.vue";

const steamStore = useSteamStore();
const { steamStatus } = storeToRefs(steamStore);
const homeData = useHomeDataStore();
const { loaded } = storeToRefs(homeData);

let refreshInterval;
onMounted(() => {
  refreshInterval = setInterval(() => steamStore.fetchSteam(), 5 * 60 * 1000);
});
onUnmounted(() => clearInterval(refreshInterval));

function formatHours(minutes) {
  const hrs = (minutes / 60).toFixed(1);
  return `${hrs}h`;
}
</script>

<template>
  <div class="flex flex-col min-h-0 overflow-hidden">
    <Header class="text-left">
      <span class="flex items-center gap-2">
        Steam
        <span
          class="inline-block w-2 h-2 rounded-full"
          :class="steamStatus.online ? 'bg-green-500' : 'bg-gray-400'"
          :title="steamStatus.online ? 'Online' : 'Offline'"
        />
      </span>
    </Header>

    <div v-if="!loaded" class="p-2 text-sm">Loading...</div>

    <div
      v-else-if="steamStatus.recentGames.length"
      class="flex-1 overflow-y-auto flex flex-col gap-2 p-1"
    >
      <div
        v-for="game in steamStatus.recentGames"
        :key="game.appId"
        class="flex flex-col"
      >
        <img
          :src="game.headerImageUrl"
          :alt="game.name"
          class="w-full object-cover"
          loading="lazy"
        />
        <div class="px-1 py-0.5 text-xs">
          <p class="font-bold truncate">{{ game.name }}</p>
          <p class="text-tertiary">
            {{ formatHours(game.playtime2Weeks) }} last 2 weeks
          </p>
        </div>
      </div>
    </div>

    <div v-else class="p-2 text-sm">No recent games.</div>
  </div>
</template>
