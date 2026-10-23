<script setup>
/**
 * Admin form: create a categorised bookmark. Emits "done"/"cancel".
 */
import Button from "@/components/input/Button.vue";
import { useCreateForm } from "@/js/useCreateForm";


const emit = defineEmits(["done", "cancel"]);

const { values, submit, error, submitting } = useCreateForm({
  mutation: `mutation CreateBookmark($input: CreateBookmarkInput!) { createBookmark(input: $input) { id } }`,
  fields: ["category", "name", "link"],
  emit,
});
const { category, name, link } = values;
</script>

<template>
  <form class="flex flex-col" @submit.prevent="submit">
    <h1>Create Bookmark</h1>
    <input type="text" v-model="category" placeholder="Category" />
    <input type="text" v-model="name" placeholder="Name" />
    <input type="text" v-model="link" placeholder="Link" />
    <small v-if="error" class="p-1" role="alert">{{ error }}</small>
    <Button type="submit" :disabled="submitting">
      {{ submitting ? "Uploading…" : "Upload" }}
    </Button>
    <Button type="button" @click="emit('cancel')">Cancel</Button>
  </form>
</template>
