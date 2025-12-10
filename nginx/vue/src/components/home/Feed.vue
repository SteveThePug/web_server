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
    <div>
        <img src="/img/epic.jpeg" />
    </div>
</template>

<style scoped>
img {
    width: 100%;
}
</style>
