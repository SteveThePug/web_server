<script setup>
import { RouterView } from "vue-router";
import Navbar from "@/components/Navbar.vue";
import Footer from "@/components/Footer.vue";
</script>

<template>
    <RouterView />

    <Navbar class="no-print" />
    <!-- <Footer style="height: 10vh" /> -->
</template>
