<script setup>
/**
 * Admin form: create a favourite. Emits "done"/"cancel" (see CreatePost.vue).
 */
import Button from "@/components/input/Button.vue";
import { useCreateForm } from "@/js/useCreateForm";


const emit = defineEmits(["done", "cancel"]);

const { values, submit: post } = useCreateForm({
  mutation: `mutation CreateFavorite($input: CreateFavoriteInput!) { createFavorite(input: $input) { id } }`,
  fields: ["type", "name", "link"],
  optional: ["link"],
  resultKey: "createFavorite",
  emit,
});
const { type, name, link } = values;
</script>

<template>
  <div class="flex flex-col">
    <h1>Create Favorite</h1>
    <input type="text" v-model="type" placeholder="Type" @keyup.enter="post" />
    <input type="text" v-model="name" placeholder="Name" @keyup.enter="post" />
    <input type="text" v-model="link" placeholder="Link" @keyup.enter="post" />
    <Button @click="post">Upload</Button>
    <Button @click="emit('cancel')">Cancel</Button>
  </div>
</template>
