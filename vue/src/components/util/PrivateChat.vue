<script setup>
/**
 * The admins-only chat widget: ChatPanel.vue on the private channel.
 *
 * The panel is mounted for admins only, so a non-admin never holds the socket
 * on its account. The box itself is not hidden: a canvas box shows to everyone
 * unless it is marked admin-only, and an empty box would look broken.
 */
import { computed } from "vue";
import ChatPanel from "./ChatPanel.vue";
import Header from "@/components/text/Header.vue";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const isAdmin = computed(() => !!auth.user.admin);
</script>

<template>
  <ChatPanel v-if="isAdmin" is-private />
  <div v-else>
    <Header>🔒 Private Chat</Header>
    <p class="p-2 text-tertiary">Admins only.</p>
  </div>
</template>
