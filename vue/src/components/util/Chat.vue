<script setup>
/**
 * The public chat widget. All of it is ChatPanel.vue; the admins-only channel
 * is the separate PrivateChat.vue widget. This file keeps the name the canvas
 * stores ("Chat" in widgetRegistry.js) pointing at the public channel.
 */
import ChatPanel from "./ChatPanel.vue";
</script>

<template>
  <ChatPanel />
</template>
