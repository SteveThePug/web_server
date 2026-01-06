<script setup>
import { ref, onMounted } from "vue";
import axios from "axios";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const userOwnsPost = ref(false);

const post = ref(null);
const fetched = ref(false);
const leftCap = ref(false);
const rightCap = ref(false);
let posts = [];
let idx = 0;
let len = 0;

async function fetchPosts() {
    try {
        const res = await axios.get("https://www.adam-french.co.uk/api/posts");
        posts = res.data;
        len = posts.length;
        fetched.value = true;
        post.value = posts[0];

        userOwnsPost.value = post.value.author.username == auth.user.username;

        leftCap.value = true;
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
    <div v-if="fetched" class="flex-row center-content">
        <h2>{{ post.title }}</h2>
        <div>{{ post.content }}</div>
        <p>by: {{ post.author.username }}</p>
        <small
            >Created at: {{ new Date(post.createdAt).toLocaleString() }}</small
        >
        <button v-if="!leftCap" @click="prevPost">Prev</button>
        <button v-if="!rightCap" @click="nextPost">Next</button>
        <button v-if="userOwnsPost" @click="deletePost">Delete</button>
    </div>
    <div class="flex-col pad" v-else>
        <h2>Can't fetch from the db yo</h2>
    </div>
</template>

<style scoped>
img {
    width: 100%;
}
</style>
