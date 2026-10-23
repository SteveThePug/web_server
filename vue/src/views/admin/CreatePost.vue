<script setup>
/**
 * Admin form: create a blog post. Emits "done" on success and "cancel" when
 * dismissed, so the host (the Feed widget's modal, or Admin.vue) decides what to
 * close or refresh. The submit plumbing is shared via useCreateForm().
 */
import Button from "@/components/input/Button.vue";
import { useCreateForm } from "@/js/useCreateForm";


const emit = defineEmits(["done", "cancel"]);

const { values, submit: post, error, submitting } = useCreateForm({
  mutation: `mutation CreatePost($input: CreatePostInput!) { createPost(input: $input) { id } }`,
  fields: ["title", "content"],
  resultKey: "createPost",
  emit,
});
const { title, content } = values;
</script>

<template>
  <form class="flex flex-col" @submit.prevent="post">
    <h1>Create Post</h1>
    <input type="text" v-model="title" placeholder="Title" />
    <textarea class="h-50" v-model="content" placeholder="Content"></textarea>
    <small v-if="error" class="p-1" role="alert">{{ error }}</small>
    <Button type="submit" :disabled="submitting">
      {{ submitting ? "Uploading…" : "Upload" }}
    </Button>
    <Button type="button" @click="emit('cancel')">Cancel</Button>
  </form>
</template>
