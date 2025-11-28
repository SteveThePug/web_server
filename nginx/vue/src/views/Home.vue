<script setup>
import Spotify from "@/components/Spotify.vue";
import Radio from "@/components/Radio.vue";
</script>

<template>
    <main>
        <div class="bordered-1">
            <h1>Welcome</h1>
            <img src="/img/epic.jpeg" />

            <h2>whoami?</h2>
            <p>Hi im Adam</p>
        </div>

        <div class="bordered-1">
            <h2>cv</h2>
            <RouterLink to="/cv">cv</RouterLink>
        </div>

        <div class="bordered-1">
            <h2>bookmarks</h2>
            <RouterLink to="/bookmarks">bookmarks</RouterLink>
        </div>

        <div class="bordered-1">
            <h2>Shrines</h2>
            <RouterLink to="/shrines/gto">gto</RouterLink>
            <RouterLink to="/shrines/demoman">demoman</RouterLink>
            <RouterLink to="/shrines/skipskipbenben">skipskipbenben</RouterLink>
            <RouterLink to="/shrines/evangelion">evangelion</RouterLink>
        </div>

        <div class="bordered-1">
            <Spotify class="border" />
        </div>
        <div class="bordered-1">
            <Radio />
        </div>
    </main>
</template>

<style scoped>
main {
    display: flex;
    flex-wrap: wrap;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1rem;
}
</style>
