<script setup>
/**
 * Admin form: log a consumed book/film/game. Emits "done"/"cancel".
 */
import Button from "@/components/input/Button.vue";
import { useCreateForm } from "@/js/useCreateForm";


const emit = defineEmits(["done", "cancel"]);

const { values, submit: post } = useCreateForm({
  mutation: `mutation CreateActivity($input: CreateActivityInput!) { createActivity(input: $input) { id } }`,
  fields: ["type", "name", "link"],
  optional: ["link"],
  resultKey: "createActivity",
  emit,
});
const { type, name, link } = values;
</script>

<template>
  <div class="flex flex-col">
    <h1>Create Activity</h1>
    <input type="text" v-model="type" placeholder="Type" @keyup.enter="post" />
    <input type="text" v-model="name" placeholder="Name" @keyup.enter="post" />
    <input type="text" v-model="link" placeholder="Link" @keyup.enter="post" />
    <Button @click="post">Upload</Button>
    <Button @click="emit('cancel')">Cancel</Button>
  </div>
</template>
