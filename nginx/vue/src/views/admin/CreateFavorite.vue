<script setup>
import Button from "@/components/input/Button.vue";

import { ref } from "vue";
import axios from "axios";

const type = ref("");
const name = ref("");
const link = ref("");

async function post() {
    try {
        const res = await axios.post("/api/favorites", {
            type: type.value,
            name: name.value,
            link: link.value || undefined,
        });
        type.value = "";
        name.value = "";
        link.value = "";
        console.log(res.data);
    } catch (err) {
        console.error(err);
    }
}
</script>

<template>
    <div class="flex flex-col">
        <h1>Create Favorite</h1>
        <input type="text" v-model="type" placeholder="Type" @keyup.enter="post" />
        <input type="text" v-model="name" placeholder="Name" @keyup.enter="post" />
        <input type="text" v-model="link" placeholder="Link" @keyup.enter="post" />
        <Button @click="post">Upload</Button>
    </div>
</template>
