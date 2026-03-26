<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import { storeToRefs } from "pinia";
import { useSteamStore } from "@/stores/steam";
import { useHomeDataStore } from "@/stores/homeData";
import Header from "@/components/text/Header.vue";

const steamStore = useSteamStore();
const { steamStatus } = storeToRefs(steamStore);
const homeData = useHomeDataStore();
const { loaded } = storeToRefs(homeData);

const idx = ref(0);
const game = computed(() => steamStatus.value.recentGames[idx.value]);

let nextId = null;
let refreshId = null;

function nextGame() {
    clearTimeout(nextId);
    nextId = setTimeout(nextGame, 5000);
    if (steamStatus.value.recentGames.length) {
        idx.value = (idx.value + 1) % steamStatus.value.recentGames.length;
    }
}

onMounted(() => {
    nextId = setTimeout(nextGame, 5000);
    refreshId = setInterval(() => steamStore.fetchSteam(), 5 * 60 * 1000);
});
onUnmounted(() => {
    clearTimeout(nextId);
    clearInterval(refreshId);
});

function formatHours(minutes) {
    const hrs = (minutes / 60).toFixed(1);
    return `${hrs}h`;
}
</script>

<template>
    <div class="steam-wrapper">
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

        <div v-else-if="game" class="game-container">
            <Transition name="fade">
                <div
                    @click="nextGame"
                    :key="game.appId"
                    class="flex flex-col items-center"
                >
                    <img
                        :src="game.headerImageUrl"
                        :alt="game.name"
                        width="145"
                        height="68"
                        class="game-img"
                    />
                    <p class="text-center">
                        <strong>{{ game.name }}</strong>
                    </p>
                    <p class="text-center text-tertiary text-xs">
                        {{ formatHours(game.playtime2Weeks) }} last 2 weeks
                    </p>
                </div>
            </Transition>
        </div>

        <div v-else class="p-2 text-sm">No recent games.</div>
    </div>
</template>

<style scoped>
.steam-wrapper {
    position: relative;
    height: 50mm;
    display: flex;
    flex-direction: column;
}

.game-container {
    position: relative;
    flex: 1;
    min-height: 0;
    overflow-y: scroll;
}

.game-img {
    width: 90%;
}

p {
    width: 100%;
    margin: 0 auto;
}

.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.5s ease;
}
.fade-leave-active {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
}
.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}

@media (max-width: 850px) {
    .steam-wrapper {
        height: auto;
        min-height: 120px;
    }
}
</style>
