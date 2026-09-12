<script setup>
import { computed, ref, defineAsyncComponent } from "vue";
import LinkTable from "@/components/util/LinkTable.vue";
import Modal from "@/components/util/Modal.vue";
import Header from "@/components/text/Header.vue";
import CreateToggle from "@/components/input/CreateToggle.vue";
import { useHomeDataStore } from "@/stores/homeData";

const CreateBookmark = defineAsyncComponent(
  () => import("@/views/admin/CreateBookmark.vue"),
);

const homeData = useHomeDataStore();

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
      Bookmarks
      <template #action><CreateToggle v-model="showCreate" /></template>
    </Header>
    <div class="bookmarks-scroll">
      <LinkTable
        v-for="group in groupedBookmarks"
        :key="group[0]"
        :title="group[0]"
        :items="group[1]"
      />
    </div>
    <Modal v-model="showCreate">
      <CreateBookmark
        @done="showCreate = false"
        @cancel="showCreate = false"
      />
    </Modal>
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
