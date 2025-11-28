<script setup>
import { RouterView } from "vue-router";
import Navbar from "@/components/Navbar.vue";
import Footer from "@/components/Footer.vue";
</script>

<template>
    <Navbar style="height: 10vh" />

    <RouterView :class="$route.name" />

    <Footer style="height: 10vh" />
</template>

<style>
.home {
    height: 80vh; /* takes full window height */
    overflow: hidden; /* removes scrolling */
    display: flex;
    flex-direction: column;
}
</style>
