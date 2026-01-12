<script setup>
import Timer from "@/components/quick/Timer.vue";
import Time from "@/components/quick/Time.vue";

import Intro from "@/components/home/Intro.vue";
import Stamps from "@/components/home/Stamps.vue";
import Listening from "@/components/home/Listening.vue";
import Links from "@/components/home/Links.vue";
import Feed from "@/components/home/Feed.vue";
import Collage from "@/components/home/Collage.vue";
import Favorites from "@/components/home/Favorites.vue";
import Gym from "@/components/home/Gym.vue";
import Watching from "@/components/home/Watching.vue";
</script>

<template>
    <div class="overlay">
        <img src="/img/memes/fire-woman.gif" class="br w-10vw" />
        <div class="flex-col tr">
            <Time class="bdr-2 bg-primary" />
            <Timer class="bdr-2 bg-primary" />
        </div>
    </div>
    <div class="background halftone" />

    <main class="center-content flex-col">
        <div class="a4page-portrait bdr-1 grid relative">
            <Intro class="intro bdr-primary bg-primary" />
            <Listening class="listening bdr-primary bg-primary" />
            <Stamps class="stamps bdr-primary bg-primary" />
            <Feed class="feed bdr-primary bg-primary" />
            <Links class="links bdr-primary bg-primary" />
            <Collage class="collage bdr-primary bg-primary" />
            <Watching class="watching bdr-primary bg-primary" />
            <Favorites class="favorites bdr-primary bg-primary" />
            <Gym class="gym bdr-primary bg-primary" />
        </div>
    </main>
</template>

<style scoped>
.grid {
    display: grid;
    grid-gap: 5px;
    grid-template-columns: repeat(10, 1fr);
    grid-template-rows: repeat(10, 1fr);
}

@media (max-width: 850px) {
    .grid {
        width: 100%;
        display: flex;
        flex-direction: column;
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

.watching {
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
