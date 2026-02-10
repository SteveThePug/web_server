<script setup>
import Timer from "@/components/util/Timer.vue";
import Elle from "@/components/elle/Elle.vue";
import Time from "@/components/util/Time.vue";
import Chat from "@/components/util/Chat.vue";
import MusicPlayer from "@/components/util/MusicPlayer.vue";

import Intro from "./Intro.vue";
import Stamps from "./Stamps.vue";
import Listening from "./Listening.vue";
import Links from "./Links.vue";
import Feed from "./Feed.vue";
import Collage from "./Collage.vue";
import Favorites from "./Favorites.vue";
import Gym from "./Gym.vue";
import Consumption from "./Consumption.vue";

import UtenaFrame from "@/components/borders/UtenaFrame.vue";
</script>

<template>
    <main class="halftone justify-center flex flex-row w-full h-full">
        <div class="h-fit flex flex-row">
            <div class="a4page-portrait homeGrid relative bdr-1">
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
                class="sidebar border-quaternary place-content-between flex-1 flex flex-col m-10 w-60 border-2"
            >
                <div class="flex flex-col flex-1">
                    <Time class="bg-bg_primary border-primary border-b" />
                    <Timer class="border-primary border-b bg-bg_primary" />
                    <Elle class="flex-1" />
                    <!-- <Chat class="bdr-2 bg-bg_primary" /> -->
                    <!-- <MusicPlayer /> -->
                </div>
                <div>
                    <img
                        src="/img/memes/fire-woman.gif"
                        class="border-tertiary border"
                    />
                </div>
            </div>
        </div>
    </main>
</template>

<style scoped>
.homeGrid > * {
    border: 2px solid var(--quaternary);
    border-color: var(--quaternary);
    background-color: var(--bg_primary);
}

.homeGrid {
    display: grid;
    grid-gap: 5px;
    grid-template-columns: repeat(10, 1fr);
    grid-template-rows: repeat(10, 1fr);
}

@media (max-width: 850px) {
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
