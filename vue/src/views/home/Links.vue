<script setup>
import RouterTable from "@/components/util/RouterTable.vue";
import LinkTable from "@/components/util/LinkTable.vue";
import Header from "@/components/text/Header.vue";

const site_links = [
    { name: "CV", link: "/cv" },
    { name: "Bookmarks", link: "/bookmarks" },
    { name: "Notes", link: "/notes/" },
    { name: "Admin", link: "/admin" },
    // { name: "Shrines", link: "/shrines" },
];

const social_links = [
    { name: "Gitea", link: "/gitea/explore/repos" },
    { name: "Steam", link: "https://steamcommunity.com/id/SteveThePug" },
    { name: "Github", link: "https://github.com/SteveThePug" },
    { name: "Spotify", link: "https://open.spotify.com/user/stevethepug" },
];
</script>

<template>
    <div class="flex flex-col overflow-auto">
        <Header>Links</Header>
        <div class="flex flex-col justify-between flex-1">
            <div class="flex flex-col gap-1">
                <RouterTable :linkArr="site_links" />
            </div>
            <div class="flex flex-col gap-1">
                <LinkTable :items="social_links" />
            </div>
        </div>
    </div>
</template>
