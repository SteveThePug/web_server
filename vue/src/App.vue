<script setup>
import { RouterView } from "vue-router";
import Navbar from "@/components/Navbar.vue";
import Footer from "@/components/Footer.vue";
</script>

<template>
    <div class="app-layout">
        <Navbar class="no-print sticky top-0 z-50" />
        <main class="app-content">
            <RouterView v-slot="{ Component }">
                <Transition name="slide" mode="out-in">
                    <component :is="Component" :key="$route.path" />
                </Transition>
            </RouterView>
        </main>
        <Footer class="sticky bottom-0 z-50" />
    </div>
</template>

<style scoped>
.app-layout {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
}

.app-content {
    flex: 1;
    overflow-y: auto;
}
</style>
