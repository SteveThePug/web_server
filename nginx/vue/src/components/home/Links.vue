<script setup>
import RouterTable from "@/components/quick/RouterTable.vue";
import LinkTable from "@/components/quick/LinkTable.vue";
import Markdown from "@/components/quick/Markdown.vue";

const site_links = [
    { name: "CV", link: "/cv" },
    { name: "Bookmarks", link: "/bookmarks" },
    { name: "Notes", link: "/notes" },
    { name: "Admin", link: "/admin" },
    { name: "Shrines", link: "/shrines" },
];

const social_links = [
    { name: "Steam", link: "https://steamcommunity.com/id/SteveThePug" },
    { name: "Github", link: "https://github.com/SteveThePug" },
    { name: "Spotify", link: "https://open.spotify.com/user/stevethepug" },
];
</script>

<template>
    <div class="flex-col space-between">
        <div class="flex-col gap">
            <RouterTable :linkArr="site_links" />
        </div>
        <div class="flex-col gap">
            <LinkTable :linkArr="social_links" />
        </div>
    </div>
</template>
