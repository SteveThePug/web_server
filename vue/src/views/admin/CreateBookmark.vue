<script setup>
import Button from "@/components/input/Button.vue";
import { ref } from "vue";
import { gql } from "@/graphql";

const emit = defineEmits(["done", "cancel"]);

const category = ref("");
const name = ref("");
const link = ref("");

async function submit() {
    try {
        await gql(
            `mutation CreateBookmark($input: CreateBookmarkInput!) { createBookmark(input: $input) { id } }`,
            { input: { category: category.value, name: name.value, link: link.value } },
        );
        category.value = "";
        name.value = "";
        link.value = "";
        emit("done");
    } catch (err) {
        console.error(err);
    }
}
</script>

<template>
    <div class="flex flex-col">
        <h1>Create Bookmark</h1>
        <input type="text" v-model="category" placeholder="Category" />
        <input type="text" v-model="name" placeholder="Name" @keyup.enter="submit" />
        <input type="text" v-model="link" placeholder="Link" @keyup.enter="submit" />
        <Button @click="submit">Upload</Button>
        <Button @click="emit('cancel')">Cancel</Button>
    </div>
</template>
