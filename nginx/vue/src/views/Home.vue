<script setup></script>

<template>
    <main>
        <h1>Welcome</h1>

        <h2>whoami?</h2>
        <p>Hi im Adam</p>

        <h2>cv</h2>
        <RouterLink to="/cv">CV</RouterLink>

        <h2>bookmarks</h2>
        <a href="/pages/bookmarks.html">bookmarks</a>

        <CurrentlyListening />

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

<script>
import CurrentlyListening from "@/components/CurrentlyListening.vue";
</script>
