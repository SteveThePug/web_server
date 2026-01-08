<script setup>
import VideoTable from "@/components/quick/VideoTable.vue";

const videoSources = [
    { name: "demoman", link: "/img/demoman/1760582395316219.webm" },
    { name: "demoman", link: "/img/demoman/1761052136609718.webm" },
    { name: "demoman", link: "/img/demoman/1761088452011210.mp4" },
    { name: "demoman", link: "/img/demoman/1761570214170465.webm" },
    { name: "demoman", link: "/img/demoman/1761828457509465.webm" },
];
</script>

<template>
    <main class="center-content flex-col">
        <div
            class="a4page-portrait bdr-1 flex-row relative scroll-y center-content"
        >
            <p>
                <a href="https://wiki.teamfortress.com/wiki/Demoman"
                    >The goat</a
                >
            </p>
            <div>
                <VideoTable :sourceArr="videoSources" />
            </div>
        </div>
    </main>
</template>
