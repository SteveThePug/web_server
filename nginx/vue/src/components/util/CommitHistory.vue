<script setup>
import axios from "axios";
import { ref, onMounted } from "vue";
import Header from "@/components/text/Header.vue";

const url = "/gitea/api/v1/users/adamf/activities/feeds?limit=1";

const feed = ref(null);
const isLoading = ref(true);
const hasError = ref(false);

async function checkFeed() {
    try {
        const res = await axios.get(url);
        feed.value = res.data[0] || null;
        hasError.value = false;
    } catch (err) {
        hasError.value = true;
    } finally {
        isLoading.value = false;
    }
}

onMounted(() => {
    checkFeed();
});
</script>

<template>
    <div class="justify-center text-center">
        <Header class="text-left">Commits</Header>

        <div v-if="isLoading">
            <p>Loading latest activity...</p>
        </div>

        <div v-else-if="hasError">
            <p>Could not fetch feed. Please try again later.</p>
        </div>

        <div
            v-else-if="feed"
            class="flex-col justify-center flex overflow-y-scroll"
        >
            <h3>Last git activity</h3>
            <img
                :src="feed.act_user.avatar_url"
                alt="User avatar"
                class="avatar"
            />
            <a :href="feed.repo.html_url">
                <h3>repo: {{ feed.repo.full_name }}</h3>
            </a>
            <p>Action: {{ feed.op_type }}</p>
            <p>Message: {{ JSON.parse(feed.content).Commits[0].Message }}</p>
            <small> {{ new Date(feed.created).toLocaleString() }}</small>
        </div>

        <div v-else>
            <p>No activity found.</p>
        </div>
    </div>
</template>
