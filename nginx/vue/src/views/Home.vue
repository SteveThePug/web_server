<script setup>
import SpotifyListening from "@/components/spotify/Listening.vue";
import SpotifyRecent from "@/components/spotify/Recent.vue";
import Radio from "@/components/Radio.vue";
</script>

<template>
    <main class="flex-row">
        <div class="border1 tile1">
            <div class="border2 bg-white shadow1">
                <img src="/img/epic.jpeg" />
                <div class="margin1">
                    <h1>Welcome</h1>
                    <p>Hi im Adam, this is my website, pls don't greif</p>
                </div>
            </div>
        </div>

        <div class="border1 tile1">
            <div class="border2 bg-white shadow1">
                <!-- <video>
                    <source src="/img/memes/1761540684738196.webm" />
                </video> -->
                <img src="/img/img.png" />
                <RouterLink to="/cv"><h2>cv</h2></RouterLink>
            </div>
        </div>

        <!-- <div class="bordered-1">
            <div class="bordered-2">
                <video>
                    <source src="/img/demoman/m2-res_720p.mp4" />
                </video>
                <h2>bookmarks</h2>
                <RouterLink to="/bookmarks">bookmarks</RouterLink>
            </div>
        </div> -->

        <!-- <div class="bordered-1">
            <div class="bordered-2">
                <video>
                    <source src="/img/demoman/1761570214170465.webm" />
                </video>
                <h2>Shrines</h2>
                <RouterLink to="/shrines/gto">gto</RouterLink>
                <RouterLink to="/shrines/demoman">demoman</RouterLink>
                <RouterLink to="/shrines/skipskipbenben"
                    >skipskipbenben</RouterLink
                >
                <RouterLink to="/shrines/evangelion">evangelion</RouterLink>
            </div>
        </div> -->

        <div class="border1 tile1">
            <Radio class="border2 bg-white shadow1" />
        </div>

        <div class="border1 tile3">
            <SpotifyRecent />
        </div>

        <div class="border1 tile1">
            <SpotifyListening class="border2 bg-white shadow1" />
        </div>
    </main>
</template>
