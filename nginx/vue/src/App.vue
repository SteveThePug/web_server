<script setup>
import { RouterView } from "vue-router";
import Navbar from "@/components/Navbar.vue";
import Footer from "@/components/Footer.vue";
</script>

<template>
    <Navbar style="height: 5vh" class="no-print" />

    <RouterView />

    <!-- <Footer style="height: 10vh" /> -->
</template>
