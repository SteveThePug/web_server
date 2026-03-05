<script setup>
import Slideshow from "@/components/util/Slideshow.vue";

const images = [
    { url: "/img/miku/miku1.gif" },
    { url: "/img/miku/miku2.gif" },
    // { url: "/img/miku/miku2.png" },
];
</script>

<template>
    <Slideshow class="p-5" :images="images" :interval="10000" />
</template>
