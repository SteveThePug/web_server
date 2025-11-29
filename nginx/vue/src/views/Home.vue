<script setup>
import SpotifyListening from "@/components/spotify/Listening.vue";
import SpotifyRecent from "@/components/spotify/Recent.vue";
import Radio from "@/components/Radio.vue";
</script>

<template>
    <main>
        <div class="bordered-1">
            <div class="bordered-2">
                <img src="/img/epic.jpeg" />
                <h1>Welcome</h1>
                <p>Hi im Adam</p>
            </div>
        </div>

        <div class="bordered-1">
            <div class="bordered-2">
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

        <div class="bordered-1">
            <SpotifyListening class="border" />
        </div>

        <div class="bordered-1">
            <SpotifyRecent class="border" />
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
    flex-direction: row;
    /*align-items: center;*/
    justify-content: center;
    gap: 1rem;
}
</style>
