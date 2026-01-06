<script setup>
import Intro from "@/components/home/Intro.vue";
import Stamps from "@/components/home/Stamps.vue";
import Listening from "@/components/home/Listening.vue";
import Links from "@/components/home/Links.vue";
import Feed from "@/components/home/Feed.vue";
import Collage from "@/components/home/Collage.vue";
import Favorites from "@/components/home/Favorites.vue";
import Gym from "@/components/home/Gym.vue";
import Watching from "@/components/home/Watching.vue";

import Halftone from "@/components/backgrounds/Halftone.vue";
import Overlay from "@/components/Overlay.vue";
</script>

<template>
    <Halftone />
    <main class="center-content flex-col">
        <div class="a4page-portrait bdr-1 grid relative">
            <Overlay>
                <!-- <img src="/img/memes/star.gif" class="ov1" /> -->
                <!-- <img src="/img/memes/welcome2.png" class="ov2" /> -->
                <img src="/img/memes/pidgeon.gif" class="ov3" />
                <img src="/img/memes/no_slip.png" class="ov5" />
                <!-- <img src="/img/memes/beer.png" class="ov6" /> -->
            </Overlay>

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
    <img src="/img/memes/fire-woman.gif" class="ov4" />
</template>

<style scoped>
.grid {
    display: grid;
    grid-gap: 10px;
    grid-template-columns: repeat(10, 1fr);
    grid-template-rows: repeat(10, 1fr);
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

.ov1 {
    rotate: 10deg;
    width: 100px;
    position: absolute;
    top: -90px;
    left: 0px;
}
.ov2 {
    width: 100px;
    position: absolute;
    top: -102px;
    right: 0px;
}
.ov3 {
    width: 200px;
    position: absolute;
    bottom: -50px;
    left: -200px;
}
.ov4 {
    width: 250px;
    position: absolute;
    bottom: 0px;
    right: 0px;
}
.ov5 {
    height: 100px;
    position: absolute;
    top: 100px;
    right: -100px;
}
.ov6 {
    height: 250px;
    position: absolute;
    top: 450px;
    right: -200px;
    animation: floatY 5s ease-in-out infinite alternate;
}

@keyframes floatY {
    from {
        transform: translateY(0);
    }
    to {
        transform: translateY(-500px);
    }
}
</style>
