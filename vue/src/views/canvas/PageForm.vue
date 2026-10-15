<script setup>
/**
 * Create a canvas page: its address, title, a starting layout from
 * pageStarters.js, and whether only admins may see it. On success it opens
 * the new page and emits `created`. Used on the admin panel's Pages tab
 * (twice: any page, and a shrine under /shrines/) and on the 404 an admin sees
 * at an address that has no page yet.
 */
import { ref } from "vue";
import { useRouter } from "vue-router";
import Button from "@/components/input/Button.vue";
import { usePagesStore } from "@/stores/pages";
import { useBoxesStore } from "@/stores/boxes";
import { STARTERS } from "./pageStarters";

const props = defineProps({
  // Pre-filled address, e.g. the path of the 404 being looked at.
  slug: { type: String, default: "" },
  // Fixed leading part of the address, e.g. "shrines/" for a new shrine.
  prefix: { type: String, default: "" },
  starter: { type: String, default: "blank" },
  placeholder: { type: String, default: "address, e.g. shrines/evangelion" },
  submitLabel: { type: String, default: "Create page" },
});

const emit = defineEmits(["created"]);
const router = useRouter();
const pages = usePagesStore();

const slug = ref(props.slug);
const title = ref("");
const starter = ref(props.starter);
const adminOnly = ref(false);
const busy = ref(false);
const error = ref("");

async function submit() {
  const fullSlug = props.prefix + slug.value.trim().replace(/^\/+|\/+$/g, "");
  if (!title.value.trim()) {
    error.value = "Title is required";
    return;
  }
  error.value = "";
  busy.value = true;
  try {
    const page = await pages.create({ slug: fullSlug, title: title.value.trim(), adminOnly: adminOnly.value });
    // Page first, then its boxes: a starter that fails halfway leaves a page
    // with some boxes, which is easy to finish by hand, rather than boxes
    // under an address with no page.
    const boxes = useBoxesStore(page.slug);
    for (const box of STARTERS.find((s) => s.id === starter.value).boxes(page.title)) {
      await boxes.create(box);
    }
    // On the 404 of the address just created the push is a no-op, since the
    // route is unchanged; `created` is what tells that view to reload.
    await router.push(`/${page.slug}`);
    emit("created", page);
  } catch (err) {
    error.value = err?.message || "Failed to create the page";
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <form class="page-form flex flex-col gap-1" @submit.prevent="submit">
    <label class="flex items-center gap-1">
      <span class="text-muted">/{{ prefix }}</span>
      <input v-model="slug" type="text" :placeholder="placeholder" required class="flex-1" />
    </label>
    <input v-model="title" type="text" placeholder="Title (required)" />
    <select v-model="starter">
      <option v-for="s in STARTERS" :key="s.id" :value="s.id">Start from: {{ s.label }}</option>
    </select>
    <label class="check">
      <input v-model="adminOnly" type="checkbox" />
      Only admins can see this page
    </label>
    <Button type="submit" :disabled="busy">{{ busy ? "Creating…" : submitLabel }}</Button>
    <p v-if="error" class="text-tertiary">{{ error }}</p>
  </form>
</template>

<style scoped>
.page-form select {
  padding: 0.5rem;
  color: var(--color-primary);
  background: var(--color-surface-deep);
  border: 1px solid var(--color-primary);
}
.check {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--color-secondary);
}
.check input {
  width: auto;
}
</style>
