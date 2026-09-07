<script setup>
import { useAuthStore } from "@/stores/auth";

// "+" / "x" button, shown to admins only. Toggles a home widget between its
// normal view and its create form. Designed for <Header>'s #action slot.
const show = defineModel({ type: Boolean, default: false });
const auth = useAuthStore();
</script>

<template>
  <button v-if="auth.user.admin" class="text-sm px-1" @click="show = !show">
    {{ show ? "x" : "+" }}
  </button>
</template>
