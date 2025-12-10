<script setup>
import { ref, onMounted } from "vue";
import axios from "axios";

const posts = ref([]);
async function fetchPosts() {
    try {
        const res = await axios.get("https://adam-french.co.uk/api/posts");
        posts.value = res.data;
    } catch (err) {
        console.error(err);
    }
}

onMounted(() => {
    fetchPosts();
});
</script>

<template>
    <div v-for="post in posts">
        <h2>{{ post.title }}</h2>
        <p>By: {{ post.author.username }}</p>
        <div>{{ post.content }}</div>
        <small
            >Created at: {{ new Date(post.CreatedAt).toLocaleString() }}</small
        >
    </div>
</template>

<style scoped>
img {
    width: 100%;
}
</style>
