<script setup>
import Markdown from "@/components/quick/Markdown.vue";

import { ref, computed, onMounted } from "vue";
import { useAuthStore } from "@/stores/auth";
import { usePostsStore } from "@/stores/posts";

const authStore = useAuthStore();
const postsStore = usePostsStore();

const idx = ref(0);

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

onMounted(() => {
    postsStore.fetchPosts();
});
</script>

<template>
    <div class="flex-col pad scroll-y left-content">
        <h2>{{ post.title }}</h2>
        <Markdown class="fill wrap" :source="post.content" />
        <p>by: {{ post.author.username }}</p>
        <small
            >Created at: {{ new Date(post.createdAt).toLocaleString() }}</small
        >
        <div class="flex-row fill-width">
            <button class="fill" v-if="!leftCap" @click="prevPost">Prev</button>
            <button class="fill" v-if="!rightCap" @click="nextPost">
                Next
            </button>
        </div>
        <button v-if="userOwnsPost" @click="deletePost">Delete</button>
    </div>
</template>

<style scoped>
img {
    width: 100%;
}
</style>
