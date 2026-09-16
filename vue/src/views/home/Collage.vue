<script setup>
/**
 * "鸟" widget on /stp — a two-tab cell: a <Slideshow> and, once signed in, the
 * Places list.
 *
 * The Places tab is added to `tabs` only when logged in, and the watcher forces
 * the view back to the collage on logout so a signed-out user is never left
 * staring at an empty private tab.
 */
import { ref, computed, watch } from "vue";
import Slideshow from "@/components/util/Slideshow.vue";
import Modal from "@/components/util/Modal.vue";
import PlacesTable from "./PlacesTable.vue";
import CreatePlace from "@/views/admin/CreatePlace.vue";
import { useAuthStore } from "@/stores/auth";

const images = [{ url: "/img/memes/pidgeon.gif", comment: "鸟" }];

const auth = useAuthStore();
const loggedIn = computed(() => auth.loggedIn);

const tabs = computed(() => [
  { id: "collage", label: "鸟" },
  ...(loggedIn.value ? [{ id: "places", label: "Places" }] : []),
]);

const activeTab = ref("collage");
const showAdd = ref(false);

// On logout the Places tab disappears from `tabs`; without this the component
// would be left pointing at a tab that no longer exists and render nothing.
watch(loggedIn, (isIn) => {
  if (!isIn && activeTab.value === "places") activeTab.value = "collage";
});
</script>

<template>
  <div class="collage-cell">
    <div class="tabs">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        class="tab"
        :class="{ 'is-active': activeTab === tab.id }"
        @click="activeTab = tab.id"
      >
        {{ tab.label }}
      </button>
      <div class="tab-spacer" />
      <button
        v-if="loggedIn && activeTab === 'places'"
        class="tab add-btn"
        title="Add a place / thing to do"
        @click="showAdd = true"
      >
        +
      </button>
    </div>

    <div class="tab-body">
      <Slideshow v-if="activeTab === 'collage'" :images="images" />
      <PlacesTable v-else-if="loggedIn" />
    </div>

    <Modal v-model="showAdd">
      <CreatePlace @done="showAdd = false" @cancel="showAdd = false" />
    </Modal>
  </div>
</template>

<style scoped>
.collage-cell {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.tabs {
  display: flex;
  align-items: stretch;
  gap: 2px;
  padding: 2px;
  border-bottom: 1px solid var(--color-quaternary);
  background-color: var(--color-surface-deep);
}

.tab {
  background-color: var(--color-link-bg);
  color: var(--color-primary);
  border: 1px solid transparent;
  padding: 2px 8px;
  font-family: var(--font-heading);
  letter-spacing: 0.05em;
  cursor: pointer;
  transition:
    background-color 120ms ease,
    border-color 120ms ease,
    color 120ms ease;
}
.tab:hover {
  border-color: var(--color-primary);
}
.tab.is-active {
  border-color: var(--color-primary);
  color: var(--color-tertiary);
  background-color: var(--color-surface-tint);
}

.tab-spacer {
  flex: 1;
}

.add-btn {
  min-width: 24px;
  padding: 0 6px;
}

.tab-body {
  flex: 1;
  min-height: 0;
  position: relative;
  overflow: hidden;
}
</style>
