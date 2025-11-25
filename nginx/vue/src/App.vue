<script setup>
import { RouterLink, RouterView } from "vue-router";
</script>

<template>
    <header>
        <div class="wrapper">
            <nav class="no-print">
                <RouterLink to="/">Home</RouterLink>
                <RouterLink to="/cv">CV</RouterLink>
            </nav>
        </div>
    </header>

    <RouterView />
</template>

<style scoped></style>
