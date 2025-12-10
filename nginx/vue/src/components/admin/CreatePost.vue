<script setup>
import { ref, onMounted, computed } from "vue";
import axios from "axios";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const loggedIn = auth.loggedIn;
const title = ref("");
const content = ref("");

async function post() {
    try {
        const res = await axios.post("/api/posts", {
            title: title.value,
            content: content.value,
        });
        title.value = "";
        content.value = "";
        console.log(res.data);
    } catch (err) {
        console.error(err);
    }
}
</script>

<template>
    <div v-if="loggedIn">
        <h1>Create Post</h1>
        <input type="text" v-model="title" placeholder="Title" />
        <textarea v-model="content" placeholder="Content"></textarea>
        <button @click="post">Upload</button>
    </div>
</template>
