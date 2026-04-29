<script setup>
import VideoTable from "@/components/util/VideoTable.vue";
import Link from "@/components/text/Link.vue";

const videoSources = [
  { name: "demoman", link: "/img/demoman/1760582395316219.webm" },
  { name: "demoman", link: "/img/demoman/1761052136609718.webm" },
  { name: "demoman", link: "/img/demoman/1761088452011210.mp4" },
  { name: "demoman", link: "/img/demoman/1761570214170465.webm" },
  { name: "demoman", link: "/img/demoman/1761828457509465.webm" },
];
</script>

<template>
  <main class="items-center flex flex-col">
    <div
      class="a4page-portrait bdr-1 flex flex-row relative overflow-scroll items-center"
    >
      <p>
        <Link href="https://wiki.teamfortress.com/wiki/Demoman">The goat</Link>
      </p>
      <div>
        <VideoTable :sourceArr="videoSources" />
      </div>
    </div>
  </main>
</template>
