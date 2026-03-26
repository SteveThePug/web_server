<script setup>
import { useHomeDataStore } from "@/stores/homeData";
import { storeToRefs } from "pinia";
import Header from "@/components/text/Header.vue";
import Link from "@/components/text/Link.vue";

const homeData = useHomeDataStore();
const { gitFeed: feed, loaded } = storeToRefs(homeData);
</script>

<template>
    <div class="flex flex-col text-center min-h-0 h-full overflow-x-hidden">
        <Header class="text-left">Commits</Header>

        <div v-if="!loaded" class="flex-1 overflow-y-auto">
            <p>Loading latest activity...</p>
        </div>

        <div
            v-else-if="feed"
            class="flex-1 flex flex-col overflow-y-auto overflow-x-hidden"
        >
            <h3>Last git activity</h3>
            <img :src="feed.avatarUrl" alt="User avatar" class="avatar" loading="lazy" />
            <Link :href="feed.repoUrl">
                <h3>repo: {{ feed.repoName }}</h3>
            </Link>
            <p>Action: {{ feed.opType }}</p>
            <p>Message: {{ feed.commitMessage }}</p>
            <small> {{ new Date(feed.createdAt).toLocaleString() }}</small>
        </div>

        <div v-else class="flex-1 overflow-y-auto">
            <p>No activity found.</p>
        </div>
    </div>
</template>
