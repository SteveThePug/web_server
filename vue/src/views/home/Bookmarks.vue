<script setup>
import { computed, ref, defineAsyncComponent } from "vue";
import LinkTable from "@/components/util/LinkTable.vue";
import Header from "@/components/text/Header.vue";
import { useHomeDataStore } from "@/stores/homeData";
import { useAuthStore } from "@/stores/auth";

const CreateBookmark = defineAsyncComponent(
  () => import("@/views/admin/CreateBookmark.vue"),
);

const homeData = useHomeDataStore();
const authStore = useAuthStore();

const showCreate = ref(false);

const groupedBookmarks = computed(() => {
  const groups = {};
  for (const b of homeData.bookmarks) {
    if (!groups[b.category]) groups[b.category] = [];
    groups[b.category].push(b);
  }
  return Object.entries(groups);
});
</script>

<template>
  <div class="bookmarks-wrapper">
    <Header class="text-left">
      <span class="flex items-center justify-between w-full">
        {{ showCreate ? "Create Bookmark" : "Bookmarks" }}
        <button
          v-if="authStore.user.admin"
          class="text-sm px-1"
          @click="showCreate = !showCreate"
        >
          {{ showCreate ? "x" : "+" }}
        </button>
      </span>
    </Header>
    <CreateBookmark
      v-if="showCreate"
      class="flex-1 min-h-0 p-1"
      @done="showCreate = false"
      @cancel="showCreate = false"
    />
    <div v-if="!showCreate" class="bookmarks-scroll">
      <LinkTable
        v-for="group in groupedBookmarks"
        :key="group[0]"
        :title="group[0]"
        :items="group[1]"
      />
    </div>
  </div>
</template>

<style scoped>
.bookmarks-wrapper {
  display: flex;
  flex-direction: column;
  min-height: 0;
  flex: 1;
}

.bookmarks-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}
</style>
