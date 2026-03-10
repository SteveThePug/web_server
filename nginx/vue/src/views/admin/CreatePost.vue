<script setup>
import Button from "@/components/input/Button.vue";
import { ref } from "vue";
import axios from "axios";

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
    <div class="flex flex-col">
        <h1>Create Post</h1>
        <input type="text" v-model="title" placeholder="Title" @keyup.enter="post" />
        <textarea
            class="h-50"
            v-model="content"
            placeholder="Content"
        ></textarea>
        <Button @click="post">Upload</Button>
        <!-- make textarea take up most the space -->
    </div>
</template>
