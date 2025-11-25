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

        <h2>Listening to:</h2>
        <div
            x-data="spotifyPlayer()"
            x-init="fetchNowPlaying(); setInterval(fetchNowPlaying, 60000)"
            class="spotify-card"
        >
            <img :src="album_image" class="album-img" alt="" />
            <div class="spotify-info">
                <div x-text="song_name || 'No song playing'"></div>
                <div x-text="artist_name"></div>
                <div
                    x-text="playing ? 'Playing' : ''"
                    :class="{ playing: playing }"
                ></div>
                <a :href="song_url"></a>
            </div>
        </div>

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
