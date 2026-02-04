<script setup>
import Timer from "@/components/util/Timer.vue";
import Time from "@/components/util/Time.vue";
import Chat from "@/components/util/Chat.vue";
import MusicPlayer from "@/components/util/MusicPlayer.vue";

import Intro from "@/components/home/Intro.vue";
import Stamps from "@/components/home/Stamps.vue";
import Listening from "@/components/home/Listening.vue";
import Links from "@/components/home/Links.vue";
import Feed from "@/components/home/Feed.vue";
import Collage from "@/components/home/Collage.vue";
import Favorites from "@/components/home/Favorites.vue";
import Gym from "@/components/home/Gym.vue";
import Consumption from "@/components/home/Consumption.vue";
</script>

<template>
    <main class="halftone justify-center flex flex-row w-full h-full p-10">
        <div class="a4height flex flex-row">
            <div class="a4page-portrait bdr-1 homeGrid relative">
                <Intro class="intro" />
                <Listening class="listening" />
                <Stamps class="stamps" />
                <Feed class="feed" />
                <Links class="links" />
                <Collage class="collage" />
                <Consumption class="consumption" />
                <Favorites class="favorites" />
                <Gym class="gym" />
            </div>
            <div
                class="flex flex-col sidebar border-primary place-content-between h-full sticky m-10 w-60"
            >
                <div class="flex flex-col">
                    <Time />
                    <Timer />
                    <!-- <Chat class="bdr-2 bg-bg_primary" /> -->
                    <!-- <MusicPlayer /> -->
                </div>
                <img
                    src="/img/memes/fire-woman.gif"
                    class="border-tertiary border"
                />
            </div>
        </div>
    </main>
</template>

<style scoped>
.a4height {
    height: 297mm;
}
.homeGrid > * {
    @apply border-2 border-dotted;
    border-color: var(--primary);
    background-color: var(--bg_primary);
}

.homeGrid {
    display: grid;
    grid-gap: 5px;
    grid-template-columns: repeat(10, 1fr);
    grid-template-rows: repeat(10, 1fr);
}

@media (max-width: 800px) {
    .homeGrid {
        width: 100%;
        display: flex;
        flex-direction: column;
    }
}

@media (max-width: 1200px) {
    .tr,
    .br,
    .sidebar {
        display: none;
    }
}

.intro {
    grid-column: 1 / span 6;
    grid-row: 1 / span 4;
}

.listening {
    grid-column: 7 / span 4;
    grid-row: 1 / span 3;
}

.stamps {
    grid-column: 7 / span 4;
    grid-row: 4 / span 1;
}

.feed {
    grid-column: 1 / span 3;
    grid-row: 5 / span 4;
}

.links {
    grid-column: 4 / span 2;
    grid-row: 5 / span 4;
}

.collage {
    grid-column: 6 / span 5;
    grid-row: 5 / span 4;
}

.consumption {
    grid-column: span 4;
    grid-row: span 2;
}
.gym {
    grid-column: span 3;
    grid-row: span 2;
}

.favorites {
    grid-column: span 3;
    grid-row: span 2;
}
</style>
