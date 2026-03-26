<script setup>
import Timer from "@/components/util/Timer.vue";
import Time from "@/components/util/Time.vue";
import Radio from "@/components/util/Radio.vue";
import Elle from "@/components/elle/Elle.vue";
import Chat from "@/components/util/Chat.vue";
import MusicPlayer from "@/components/util/MusicPlayer.vue";
import CommitHistory from "@/components/util/CommitHistory.vue";

import Intro from "./Intro.vue";
import Intro2 from "./Intro2.vue";
import BadApple from "./BadApple.vue";
import Miku from "./Miku.vue";
import Stamps from "./Stamps.vue";
import Listening from "./Listening.vue";
import Links from "./Links.vue";
import Feed from "./Feed.vue";
import Collage from "./Collage.vue";
import Favorites from "./Favorites.vue";
// import Gym from "./Gym.vue";
import Gym2 from "./Gym2.vue";
import Consumption from "./Consumption.vue";
import Steam from "./Steam.vue";
</script>

<template>
    <main class="halftone justify-center flex flex-row w-full h-full">
        <div class="outerWrap flex flex-row" style="height: 310mm">
            <div class="sidebar flex-1 flex flex-col m-10 w-60 gap-2">
                <div
                    class="flex-1 flex flex-col min-h-0 background-children border-children gap-2"
                >
                    <Chat class="flex-1 min-h-0" />
                    <Steam />
                </div>
                <div class="sidebar-image">
                    <Miku class="border-tertiary border bg-bg_secondary h-60" />
                </div>
            </div>
            <div
                class="a4page-portrait homeGrid relative background-children border-children bdr-1"
            >
                <!-- <Intro class="intro" /> -->
                <Intro2 class="intro" />
                <!-- <BadApple class="intro" /> -->
                <Listening class="listening" />
                <Stamps class="stamps" />
                <Feed class="feed" />
                <Links class="links" />
                <Collage class="collage" />
                <Consumption class="consumption" />
                <Favorites class="favorites" />
                <!-- <Gym class="gym" /> -->
                <Gym2 class="gym" />
            </div>
            <div class="sidebar flex-1 flex flex-col m-10 w-60 gap-2">
                <div
                    class="flex-1 flex flex-col min-h-0 background-children border-children gap-2"
                >
                    <Time />
                    <Timer />
                    <Radio />
                    <CommitHistory class="flex-1" />

                    <!-- <Elle class="flex-1" /> -->
                    <!-- <MusicPlayer /> -->
                </div>
                <div class="sidebar-image">
                    <img
                        src="/img/memes/fire-woman.gif"
                        class="border-tertiary border"
                        loading="lazy"
                    />
                </div>
            </div>
        </div>
    </main>
</template>

<style scoped>
.border-children > * {
    border: 2px solid var(--quaternary);
}

.background-children > * {
    background-color: var(--bg_primary);
}

.homeGrid {
    display: grid;
    gap: 5px;
    grid-template-columns: repeat(10, 1fr);
    grid-template-rows: repeat(10, 1fr);
}

@media (max-width: 1200px) {
    .outerWrap {
        flex-direction: column;
        align-items: stretch;
    }

    .homeGrid {
        order: -1;
        width: 100%;
        height: 350mm;
        margin-inline: 0;
        box-sizing: border-box;
    }

    .sidebar {
        width: 100%;
        margin: 5px 10px;
        flex-direction: column;
        align-items: center;
        gap: 8px;
    }
}

@media (max-width: 850px) {
    .homeGrid {
        display: flex;
        flex-direction: column;
        height: auto;
    }

    .stamps {
        max-height: 130px;
    }

    .sidebar {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
        gap: 8px;
        justify-items: stretch;
    }

    .sidebar > * {
        max-width: none;
        width: auto;
    }

    .sidebar-image {
        max-height: 200px;
        overflow: hidden;
    }

    .sidebar-image :deep(img) {
        max-height: 200px;
        object-fit: contain;
        display: block;
        margin-inline: auto;
    }
}

@media (max-width: 500px) {
    main {
        overflow-x: hidden;
    }

    .outerWrap {
        max-width: 100vw;
    }

    .sidebar {
        margin: 5px 0;
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

.bg-random {
    background-color: var(--bg_primary);
    background-image: url("/img/miku/miku2.gif");
    background-size: 10px 10px;
}
</style>
