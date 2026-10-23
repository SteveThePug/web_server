<script setup>
/**
 * Admin form: create a place / thing to do. Goes through stores/places.js rather
 * than gql() directly, so the new row appears in the Places table immediately.
 */
import Button from "@/components/input/Button.vue";
import { ref } from "vue";
import { usePlacesStore } from "@/stores/places";

const emit = defineEmits(["done", "cancel"]);
const places = usePlacesStore();

const title = ref("");
const location = ref("");
const category = ref("");
const cost = ref("");
const priority = ref(0);
const imageUrl = ref("");
const notes = ref("");
const done = ref(false);
const error = ref("");
const submitting = ref(false);

async function submit() {
  if (submitting.value) return;
  if (!title.value.trim()) {
    error.value = "Title is required";
    return;
  }
  error.value = "";
  submitting.value = true;
  try {
    await places.create({
      title: title.value.trim(),
      location: location.value.trim() || null,
      category: category.value.trim() || null,
      cost: cost.value.trim() || null,
      priority: Number(priority.value) || 0,
      imageUrl: imageUrl.value.trim() || null,
      notes: notes.value.trim() || null,
      done: done.value,
    });
    title.value = "";
    location.value = "";
    category.value = "";
    cost.value = "";
    priority.value = 0;
    imageUrl.value = "";
    notes.value = "";
    done.value = false;
    emit("done");
  } catch (err) {
    error.value = err?.message || "Failed to save";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <form class="flex flex-col" @submit.prevent="submit">
    <h1>Add Place / Thing To Do</h1>
    <input type="text" v-model="title" placeholder="Title (required)" />
    <input type="text" v-model="location" placeholder="Location" />
    <input
      type="text"
      v-model="category"
      placeholder="Category (place, activity, food…)"
    />
    <input type="text" v-model="cost" placeholder="Cost (e.g. £20 pp)" />
    <input
      type="number"
      v-model="priority"
      placeholder="Priority (higher = sooner)"
    />
    <input type="text" v-model="imageUrl" placeholder="Image URL" />
    <textarea class="h-25" v-model="notes" placeholder="Notes"></textarea>
    <label class="flex flex-row items-center gap-2 p-2 text-primary">
      <input type="checkbox" v-model="done" class="w-auto" />
      Already done
    </label>
    <small v-if="error" class="p-1" role="alert">{{ error }}</small>
    <Button type="submit" :disabled="submitting">
      {{ submitting ? "Uploading…" : "Upload" }}
    </Button>
    <Button type="button" @click="emit('cancel')">Cancel</Button>
  </form>
</template>
