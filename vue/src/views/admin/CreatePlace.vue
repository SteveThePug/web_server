<script setup>
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

async function submit() {
  if (!title.value.trim()) {
    error.value = "Title is required";
    return;
  }
  error.value = "";
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
  }
}
</script>

<template>
  <div class="flex flex-col">
    <h1>Add Place / Thing To Do</h1>
    <input
      type="text"
      v-model="title"
      placeholder="Title (required)"
      @keyup.enter="submit"
    />
    <input
      type="text"
      v-model="location"
      placeholder="Location"
      @keyup.enter="submit"
    />
    <input
      type="text"
      v-model="category"
      placeholder="Category (place, activity, food…)"
      @keyup.enter="submit"
    />
    <input
      type="text"
      v-model="cost"
      placeholder="Cost (e.g. £20 pp)"
      @keyup.enter="submit"
    />
    <input
      type="number"
      v-model="priority"
      placeholder="Priority (higher = sooner)"
      @keyup.enter="submit"
    />
    <input
      type="text"
      v-model="imageUrl"
      placeholder="Image URL"
      @keyup.enter="submit"
    />
    <textarea class="h-25" v-model="notes" placeholder="Notes"></textarea>
    <label class="flex flex-row items-center gap-2 p-2 text-primary">
      <input type="checkbox" v-model="done" class="w-auto" />
      Already done
    </label>
    <small v-if="error" class="p-1">{{ error }}</small>
    <Button @click="submit">Upload</Button>
    <Button @click="emit('cancel')">Cancel</Button>
  </div>
</template>
