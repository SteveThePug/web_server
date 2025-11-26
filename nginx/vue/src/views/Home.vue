<script setup>
import Spotify from "@/components/Spotify.vue";
import Radio from "@/components/Radio.vue";
</script>

<template>
    <Spotify class="floating" style="left: 10vw; top: 10vh" />
    <Radio class="floating" style="left: 90vw; top: 10vh" />

    <main class="a4page">
        <h1>Welcome</h1>
        <img src="/img/epic.jpeg" />

        <h2>whoami?</h2>
        <p>Hi im Adam</p>

        <h2>cv</h2>
        <RouterLink to="/cv">cv</RouterLink>

        <h2>bookmarks</h2>
        <RouterLink to="/bookmarks">bookmarks</RouterLink>

        <div>
            <h2>Shrines</h2>
            <RouterLink to="/shrines/gto">gto</RouterLink>
            <RouterLink to="/shrine/demoman">demoman</RouterLink>
            <RouterLink to="/shrines/skipskipbenben">skipskipbenben</RouterLink>
            <RouterLink to="/shrines/evangelion">evangelion</RouterLink>
        </div>
    </main>
</template>

<styles scoped>
</styles>
