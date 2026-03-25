<script setup>
import Slideshow from "@/components/util/Slideshow.vue";

const images = [
    { url: "/img/memes/pidgeon.gif", comment: "鸟" },
    // { url: "/img/memes/no_slip.png" },
    // { url: "/img/memes/epic.jpeg" },
    // { url: "/img/bedroom/img2.png", comment: "办公桌" },
    // { url: "/img/bedroom/img1.png", comment: "床" },
];
</script>

<template>
    <Slideshow :images="images" />
</template>
