<script setup>
/**
 * Admin form: create a categorised bookmark. Emits "done"/"cancel".
 */
import Button from "@/components/input/Button.vue";
import { useCreateForm } from "@/js/useCreateForm";


const emit = defineEmits(["done", "cancel"]);

const { values, submit } = useCreateForm({
  mutation: `mutation CreateBookmark($input: CreateBookmarkInput!) { createBookmark(input: $input) { id } }`,
  fields: ["category", "name", "link"],
  emit,
});
const { category, name, link } = values;
</script>

<template>
  <div class="flex flex-col">
    <h1>Create Bookmark</h1>
    <input type="text" v-model="category" placeholder="Category" />
    <input
      type="text"
      v-model="name"
      placeholder="Name"
      @keyup.enter="submit"
    />
    <input
      type="text"
      v-model="link"
      placeholder="Link"
      @keyup.enter="submit"
    />
    <Button @click="submit">Upload</Button>
    <Button @click="emit('cancel')">Cancel</Button>
  </div>
</template>
