<script setup>
/**
 * Admin form: create a favourite. Emits "done"/"cancel" (see CreatePost.vue).
 */
import Button from "@/components/input/Button.vue";
import { useCreateForm } from "@/js/useCreateForm";


const emit = defineEmits(["done", "cancel"]);

const { values, submit: post, error, submitting } = useCreateForm({
  mutation: `mutation CreateFavorite($input: CreateFavoriteInput!) { createFavorite(input: $input) { id } }`,
  fields: ["type", "name", "link"],
  optional: ["link"],
  resultKey: "createFavorite",
  emit,
});
const { type, name, link } = values;
</script>

<template>
  <form class="flex flex-col" @submit.prevent="post">
    <h1>Create Favorite</h1>
    <input type="text" v-model="type" placeholder="Type" />
    <input type="text" v-model="name" placeholder="Name" />
    <input type="text" v-model="link" placeholder="Link" />
    <small v-if="error" class="p-1" role="alert">{{ error }}</small>
    <Button type="submit" :disabled="submitting">
      {{ submitting ? "Uploading…" : "Upload" }}
    </Button>
    <Button type="button" @click="emit('cancel')">Cancel</Button>
  </form>
</template>
