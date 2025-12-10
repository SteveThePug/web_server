<script setup>
import { ref, onMounted } from "vue";
import axios from "axios";

const post = ref(null);
const fetched = ref(false);
const leftCap = ref(false);
const rightCap = ref(false);
let posts = [];
let idx = 0;
let len = 0;

async function fetchPosts() {
    try {
        const res = await axios.get("/api/posts");
        posts = res.data;
        fetched.value = true;
        post.value = posts[0];
        leftCap.value = true;
        len = posts.length;
    } catch (err) {
        console.error(err);
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
        <button v-if="!leftCap" @click="prevPost">Prev</button>
        <button v-if="!rightCap" @click="nextPost">Next</button>
    </div>
</template>

<style scoped>
img {
    width: 100%;
}
</style>
