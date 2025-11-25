<script setup>
import Spotify from "@/components/Spotify.vue";
import Radio from "@/components/Radio.vue";
</script>

<template>
    <main class="a4page">
        <h1>Welcome</h1>

        <h2>whoami?</h2>
        <p>Hi im Adam</p>

        <h2>cv</h2>
        <RouterLink to="/cv">cv</RouterLink>

        <h2>bookmarks</h2>
        <RouterLink to="/bookmarks">bookmarks</RouterLink>

        <Spotify />

        <Radio />

        <!--<h2> </h2>
        <p>
            Sometimes there's this fire that sends shivers down my back.
            It'll come when I'm lis
        </p>
        -->

        <!--<h2>Shrines</h1>
        <a href="/pages/shrines/evangelion.html">Evangelion</a>
        <a href="/pages/shrines/skipskipbenben.html">Skip skip ben ben</a>
        <a href="/pages/shrines/demoman.html">demoman</a>-->
        <!--<a href="pages/shrines/gto.html">GTO</a>-->
    </main>
</template>
