<script setup>
import { ref, onMounted } from "vue";
import axios from "axios";

const post = ref(null);
const fetched = ref(false);
let posts = [];
let idx = 0;
let len = 0;

async function fetchPosts() {
    try {
        const res = await axios.get("/api/posts");
        posts = res.data;
        fetched.value = true;
        post.value = posts[0];
        len = posts.length;
    } catch (err) {
        console.error(err);
    }
}

function nextPost() {
    post.value = posts[idx];
    idx = (idx + 1) % len;
}

function prevPost() {
    post.value = posts[idx];
    idx = (idx - 1) % len;
}

onMounted(() => {
    fetchPosts();
});
</script>

<template>
    <div v-if="fetched">
        <h2>{{ post.title }}</h2>
        <p>By: {{ post.author.username }}</p>
        <div>{{ post.content }}</div>
        <small
            >Created at: {{ new Date(post.CreatedAt).toLocaleString() }}</small
        >
        <button @click="nextPost">Next</button>
        <button @click="prevPost">Prev</button>
    </div>
</template>

<style scoped>
img {
    width: 100%;
}
</style>
