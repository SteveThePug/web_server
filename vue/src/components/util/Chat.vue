<script setup>
/**
 * The chat widget: message list, composer, file attach, admin delete and an
 * admins-only "private" channel. Drives stores/messages.js and is mounted in
 * the right-hand sidebar of views/home/Home.vue.
 *
 * Admins see two tabs. The socket carries both channels in one stream (the
 * server only sends private messages to admin connections), so the tab is a
 * view over `messages` and also decides whether what you send is private —
 * you cannot post to a channel you are not looking at.
 *
 * Owns the socket lifecycle — connect() on mount, disconnect() on unmount — so
 * the socket exists only while the widget is on screen.
 */
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue";
import Button from "@/components/input/Button.vue";
import { useMessagesStore } from "@/stores/messages";
import { useAuthStore } from "@/stores/auth";
import Header from "@/components/text/Header.vue";
import Link from "@/components/text/Link.vue";

const messagesStore = useMessagesStore();
const authStore = useAuthStore();
const messageInput = ref("");
const messagesContainer = ref(null);
const messagesInner = ref(null);
const fileInput = ref(null);
const isAdmin = computed(() => !!authStore.user.admin);

const activeTab = ref("public");
const isPrivateTab = computed(
  () => isAdmin.value && activeTab.value === "private",
);

// The private tab is the only place private messages are shown, so the public
// tab stays public even for an admin.
const messages = computed(() =>
  messagesStore.messages.filter((m) => !!m.private === isPrivateTab.value),
);

// Autoscroll only when the reader is already at the bottom: scrolling up to
// read history must not be yanked back down by an incoming message.
const isNearBottom = ref(true);
const SCROLL_THRESHOLD = 100;
let resizeObserver = null;

function scrollToBottom() {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
  }
}

function scrollToBottomIfNear() {
  if (isNearBottom.value) {
    scrollToBottom();
  }
}

function onScroll() {
  if (!messagesContainer.value) return;
  const { scrollHeight, scrollTop, clientHeight } = messagesContainer.value;
  isNearBottom.value =
    scrollHeight - scrollTop - clientHeight < SCROLL_THRESHOLD;
}

function goToBottom() {
  isNearBottom.value = true;
  scrollToBottom();
}

// nextTick: scrollHeight is only correct once Vue has patched the new row in.
watch(
  () => messages.value.length,
  () => {
    nextTick(scrollToBottomIfNear);
  },
);

// A tab switch swaps the whole list, so start the new one at its newest message
// rather than at whatever scroll offset the old one left behind.
watch(activeTab, () => {
  isNearBottom.value = true;
  nextTick(scrollToBottom);
});

function sendMessage() {
  const text = messageInput.value.trim();
  if (!text) return;
  isNearBottom.value = true;
  messagesStore.sendMessage(text, isPrivateTab.value);
  messageInput.value = "";
}

function deleteMessage(id) {
  messagesStore.deleteMessage(id);
}

// Admin status is fixed when the socket connects (from the auth cookies), so
// reconnect after a login/logout to pick up or drop private messages. A
// demoted admin must also be moved off the private tab, which no longer has
// anything to show.
watch(isAdmin, (admin) => {
  if (!admin) activeTab.value = "public";
  if (!messagesStore.isConnected) return;
  messagesStore.disconnect();
  messagesStore.connect();
});

async function onFileSelected(e) {
  const file = e.target.files[0];
  if (!file) return;
  isNearBottom.value = true;
  // The private flag both routes the upload to /uploads/private/ and marks the
  // message itself private; the two must agree.
  await messagesStore.uploadAndSendFile(file, isPrivateTab.value);
  fileInput.value.value = "";
}

function isImageUrl(url) {
  return /\.(jpg|jpeg|png|gif|webp)$/i.test(url);
}

function isVideoUrl(url) {
  return /\.(mp4|webm|ogg|mov)$/i.test(url);
}

// Public attachments live under /uploads/; attachments on private messages are
// written under /uploads/private/, which nginx gates behind an admin check.
const UPLOAD_PREFIXES = ["/uploads/", "/uploads/private/"];

/** Only render attachments the backend served from its own upload directory —
 *  the fileUrl arrives over the socket and is otherwise attacker-controlled.
 *  A strict prefix allowlist, so absolute ("https://evil/uploads/x") and
 *  protocol-relative ("//evil/uploads/x") URLs are rejected; "..", which could
 *  otherwise escape the directory, is rejected outright. */
function isSafeFileUrl(url) {
  if (typeof url !== "string") return false;
  if (url.includes("..")) return false;
  return UPLOAD_PREFIXES.some((prefix) => url.startsWith(prefix));
}

// NOTE: a module-level /g regex carries `lastIndex` between calls. It is safe
// here only because parseMessageParts() below always runs exec() to completion
// (exec returns null, resetting lastIndex) before returning. An early return
// from that loop would make the next message parse from the wrong offset.
const urlRegex = /(https?:\/\/[^\s<]+)/g;

/** Split message text into alternating {type:'text'|'link', value} parts so
 *  URLs can be rendered as anchors without using v-html. */
function parseMessageParts(text) {
  const parts = [];
  let lastIndex = 0;
  let match;
  while ((match = urlRegex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      parts.push({
        type: "text",
        value: text.slice(lastIndex, match.index),
      });
    }
    parts.push({ type: "link", value: match[1] });
    lastIndex = urlRegex.lastIndex;
  }
  if (lastIndex < text.length) {
    parts.push({ type: "text", value: text.slice(lastIndex) });
  }
  return parts;
}

onMounted(() => {
  messagesStore.connect();

  if (messagesContainer.value) {
    messagesContainer.value.addEventListener("scroll", onScroll, {
      passive: true,
    });
  }

  // Images and videos load after their row is inserted and change its height,
  // which the length watcher above has already missed. Observing the inner
  // wrapper catches those late reflows.
  if (messagesInner.value) {
    resizeObserver = new ResizeObserver(scrollToBottomIfNear);
    resizeObserver.observe(messagesInner.value);
  }

  scrollToBottom();
});

onUnmounted(() => {
  messagesStore.disconnect();

  if (messagesContainer.value) {
    messagesContainer.value.removeEventListener("scroll", onScroll);
  }

  if (resizeObserver) {
    resizeObserver.disconnect();
    resizeObserver = null;
  }
});
</script>

<template>
  <div class="chat-root flex-col flex min-h-0">
    <Header>{{ isPrivateTab ? "Private Chat" : "Chat" }}</Header>
    <div v-if="isAdmin" class="flex gap-1 pt-1" role="tablist">
      <button
        type="button"
        class="chat-tab flex-1"
        role="tab"
        :aria-selected="activeTab === 'public'"
        :class="{ 'is-active': activeTab === 'public' }"
        @click="activeTab = 'public'"
      >
        Public
      </button>
      <button
        type="button"
        class="chat-tab flex-1"
        role="tab"
        :aria-selected="activeTab === 'private'"
        :class="{ 'is-active': activeTab === 'private' }"
        @click="activeTab = 'private'"
      >
        🔒 Private
      </button>
    </div>
    <div
      ref="messagesContainer"
      class="flex flex-col flex-1 min-h-0 overflow-y-auto overflow-x-hidden p-2 min-w-0"
    >
      <div ref="messagesInner">
        <p
          v-for="message in messages"
          :key="message.id"
          class="break-words min-w-0 w-full"
          :class="{ 'text-secondary italic': message.private }"
        >
          <button
            v-if="isAdmin && message.id"
            type="button"
            class="text-tertiary hover:text-primary cursor-pointer mr-1"
            aria-label="Delete message"
            title="Delete message"
            @click="deleteMessage(message.id)"
          >
            ×
          </button>
          <span class="text-tertiary">{{ message.authorId }}:</span>
          <template
            v-for="(part, i) in parseMessageParts(message.text || '')"
            :key="i"
          >
            <Link
              v-if="part.type === 'link'"
              bare
              :href="part.value"
              target="_blank"
              class="text-primary underline break-all"
              >{{ part.value }}</Link
            >
            <span v-else>{{ part.value }}</span>
          </template>
          <template v-if="message.fileUrl && isSafeFileUrl(message.fileUrl)">
            <img
              v-if="isImageUrl(message.fileUrl)"
              :src="message.fileUrl"
              alt="Uploaded image"
              loading="lazy"
              class="w-full max-w-full rounded block"
            />
            <video
              v-else-if="isVideoUrl(message.fileUrl)"
              :src="message.fileUrl"
              controls
              preload="none"
              class="w-full max-w-full max-h-48 rounded block"
            />
            <Link
              v-else
              bare
              :href="message.fileUrl"
              target="_blank"
              class="underline break-all"
              >{{ message.fileUrl.split("/").pop() }}</Link
            >
          </template>
        </p>
      </div>
    </div>
    <div>
      <input
        v-model="messageInput"
        @keyup.enter="sendMessage"
        :aria-label="isPrivateTab ? 'Private chat message' : 'Chat message'"
        :placeholder="isPrivateTab ? 'Admins only' : ''"
      />
      <input
        ref="fileInput"
        type="file"
        class="hidden"
        @change="onFileSelected"
      />
      <div class="flex gap-2">
        <Button class="flex-1" @click="sendMessage">Send</Button>
        <Button v-if="isAdmin" class="flex-1" @click="fileInput.click()"
          >Attach</Button
        >
        <Button v-if="!isNearBottom" class="flex-1" @click="goToBottom"
          >Bottom</Button
        >
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-tab {
  padding: 2px 6px;
  color: var(--color-primary);
  background-color: var(--color-link-bg);
  border: 1px solid transparent;
  font-family: var(--font-heading);
  font-size: 0.875rem;
  cursor: pointer;
  transition:
    background-color 120ms ease,
    border-color 120ms ease,
    color 120ms ease;
}
.chat-tab:hover {
  border-color: var(--color-primary);
}
.chat-tab.is-active {
  border-color: var(--color-primary);
  color: var(--color-tertiary);
  background-color: var(--color-surface-tint);
}

@media (max-width: 850px) {
  .chat-root {
    max-height: none;
    height: 100%;
  }
}
</style>
