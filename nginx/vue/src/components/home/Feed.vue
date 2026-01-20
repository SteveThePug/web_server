<script setup>
import Markdown from "@/components/quick/Markdown.vue";

import { ref, onMounted } from "vue";
import axios from "axios";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const userOwnsPost = ref(false);

const post = ref({
    title: "Can't fetch from the db yo",
    content:
        "This is meant to be pulling from a database, but for some reason that isn't working and this is filler text that should hopefully never see the light of day. If you are reading this, something has gone horribly, horribly wrong. Please start crying and prepare for the incoming wrath of hell. Furthermore, this is very, very long because I am trying to test the scroll feature so thank you ^_^.",
    author: {
        username: "stp",
    },
    createdAt: Date.now(),
});
const leftCap = ref(false);
const rightCap = ref(false);
let posts = [];
let idx = 0;
let len = 0;

async function fetchPosts() {
    try {
        const res = await axios.get("/api/posts");
        if (Array.isArray(res.data)) {
            posts = res.data;
            len = posts.length;
            post.value = posts[0];
            userOwnsPost.value =
                post.value.author.username == auth.user.username;
            leftCap.value = true;
        } else {
            throw new Error("Invalid response from API");
        }
    } catch (err) {
        console.log("Cannot connect to API");
    }
}

function nextPost() {
    if (idx < len - 1) {
        idx++;
        rightCap.value = idx === len - 1;
        leftCap.value = idx === 0;
        post.value = posts[idx];
    }
}

function prevPost() {
    if (idx > 0) {
        idx--;
        rightCap.value = idx === len - 1;
        leftCap.value = idx === 0;
        post.value = posts[idx];
    }
}

async function deletePost() {
    try {
        const res = await axios.delete(
            `/api/posts/${encodeURIComponent(post.value.id)}`,
        );
        console.log("Deleted:", res.data);
        fetchPosts();
    } catch (err) {
        console.error("Delete failed:", err);
    }
}

onMounted(() => {
    fetchPosts();
});
</script>

<template>
    <div class="flex-col pad scroll-y left-content">
        <h2>{{ post.title }}</h2>
        <Markdown class="fill wrap" :source="post.content" />
        <p>by: {{ post.author.username }}</p>
        <small
            >Created at: {{ new Date(post.createdAt).toLocaleString() }}</small
        >
        <div class="flex-row fill-width">
            <button class="fill" v-if="!leftCap" @click="prevPost">Prev</button>
            <button class="fill" v-if="!rightCap" @click="nextPost">
                Next
            </button>
        </div>
        <button v-if="userOwnsPost" @click="deletePost">Delete</button>
    </div>
</template>

<style scoped>
img {
    width: 100%;
}
</style>
