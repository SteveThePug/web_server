<script setup>
/**
 * Shared shape of the "favs" and "Consumption" widgets on /stp: a header with a
 * create toggle, a self-scrolling <AutoScroll> list rendered through
 * <LinkTable>, and an admin-only create form in a <Modal>.
 *
 * `createForm` is the component to show in the modal. The hosts pass a
 * defineAsyncComponent() so the admin form is still only fetched when a widget
 * is actually opened.
 */
import Header from "@/components/text/Header.vue";
import CreateToggle from "@/components/input/CreateToggle.vue";
import LinkTable from "@/components/util/LinkTable.vue";
import AutoScroll from "@/components/util/AutoScroll.vue";
import Modal from "@/components/util/Modal.vue";

import { ref } from "vue";

defineProps({
  title: { type: String, required: true },
  items: { type: Array, required: true },
  createForm: { type: [Object, Function], required: true },
});

const showCreate = ref(false);
</script>

<template>
  <div class="flex flex-col items-center">
    <Header>
      {{ title }}
      <template #action><CreateToggle v-model="showCreate" /></template>
    </Header>
    <AutoScroll class="flex-1 w-full">
      <LinkTable variant="table" class="w-full" :items="items" />
    </AutoScroll>
    <Modal v-model="showCreate">
      <component
        :is="createForm"
        @done="showCreate = false"
        @cancel="showCreate = false"
      />
    </Modal>
  </div>
</template>
