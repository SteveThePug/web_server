<script setup>
import Button from "@/components/input/Button.vue";
import Markdown from "@/components/util/Markdown.vue";
import Header from "@/components/text/Header.vue";

import { ref, computed, defineAsyncComponent } from "vue";
import { useAuthStore } from "@/stores/auth";
import { usePostsStore } from "@/stores/posts";

const CreatePost = defineAsyncComponent(
  () => import("@/views/admin/CreatePost.vue"),
);

const authStore = useAuthStore();
const postsStore = usePostsStore();

const idx = ref(0);
const showCreate = ref(false);

const leftCap = computed(() => idx.value === 0);
const rightCap = computed(() => idx.value === postsStore.postsCount - 1);

const post = computed(() => postsStore.posts[idx.value]);
const userOwnsPost = computed(
  () => post.value.author.username == authStore.user.username,
);

function nextPost() {
  if (idx.value < postsStore.postsCount - 1) {
    idx.value++;
  }
}

function prevPost() {
  if (idx.value > 0) {
    idx.value--;
  }
}

function deletePost() {
  postsStore.deletePost(post.value);
}
</script>

<template>
  <div class="flex flex-col flex-1 min-h-0">
    <Header>
      <span class="flex items-center justify-between w-full">
        {{ showCreate ? "Create Post" : post.title }}
        <button
          v-if="authStore.user.admin"
          class="text-sm px-1"
          @click="showCreate = !showCreate"
        >
          {{ showCreate ? "x" : "+" }}
        </button>
      </span>
    </Header>
    <CreatePost
      v-if="showCreate"
      class="flex-1 min-h-0 p-1"
      @done="showCreate = false"
      @cancel="showCreate = false"
    />
    <div
      v-if="!showCreate"
      class="flex flex-col flex-1 min-h-0 p-1 overflow-auto text-left items-start justify-start"
    >
      <small>Created at: {{ new Date(post.createdAt).toLocaleString() }}</small>
      <small>By: {{ post.author.username }}</small>
      <Markdown class="flex-1 border-box text-wrap" :source="post.content" />
      <div class="flex flex-row w-full">
        <Button class="flex-1 border-box" v-if="!leftCap" @click="prevPost">
          Prev
        </Button>
        <Button class="flex-1 border-box" v-if="!rightCap" @click="nextPost">
          Next
        </Button>
      </div>
      <Button class="w-full" v-if="userOwnsPost" @click="deletePost"
        >Delete</Button
      >
    </div>
  </div>
</template>

<style scoped>
img {
  width: 100%;
}
</style>
