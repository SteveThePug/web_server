<script setup>
/**
 * Admin form: create a blog post. Emits "done" on success and "cancel" when
 * dismissed, so the host (the Feed widget's modal, or Admin.vue) decides what to
 * close or refresh. The submit plumbing is shared via useCreateForm().
 */
import Button from "@/components/input/Button.vue";
import { useCreateForm } from "@/js/useCreateForm";


const emit = defineEmits(["done", "cancel"]);

const { values, submit: post } = useCreateForm({
  mutation: `mutation CreatePost($input: CreatePostInput!) { createPost(input: $input) { id } }`,
  fields: ["title", "content"],
  resultKey: "createPost",
  emit,
});
const { title, content } = values;
</script>

<template>
  <div class="flex flex-col">
    <h1>Create Post</h1>
    <input
      type="text"
      v-model="title"
      placeholder="Title"
      @keyup.enter="post"
    />
    <textarea class="h-50" v-model="content" placeholder="Content"></textarea>
    <Button @click="post">Upload</Button>
    <Button @click="emit('cancel')">Cancel</Button>
  </div>
</template>
