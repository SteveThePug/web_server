<script setup>
import { RouterView } from "vue-router";
import Navbar from "@/components/Navbar.vue";
import Footer from "@/components/Footer.vue";
</script>

<template>
    <Navbar class="no-print sticky" />
    <RouterView v-slot="{ Component }">
        <Transition name="slide" mode="out-in">
            <component :is="Component" :key="$route.path" />
        </Transition>
    </RouterView>

    <!-- <Footer style="height: 10vh" /> -->
</template>
