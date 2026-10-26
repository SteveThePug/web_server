<script setup>
/**
 * Places-to-go widget: the <PlacesTable> under a header whose "+" opens the
 * create form. Sign-in only, like the places mutations themselves, so a
 * signed-out visitor gets a one-line prompt instead of an empty table.
 *
 * Its own "+" rather than <CreateToggle>, because that one is admin-only and
 * any signed-in user may add a place.
 */
import { computed, ref } from "vue";
import Header from "@/components/text/Header.vue";
import Modal from "@/components/util/Modal.vue";
import PlacesTable from "./PlacesTable.vue";
import CreatePlace from "@/views/admin/CreatePlace.vue";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const loggedIn = computed(() => auth.loggedIn);
const showAdd = ref(false);
</script>

<template>
  <div class="places-cell">
    <Header class="text-left">
      Places
      <template v-if="loggedIn" #action>
        <button
          class="text-sm px-1"
          title="Add a place / thing to do"
          @click="showAdd = !showAdd"
        >
          {{ showAdd ? "x" : "+" }}
        </button>
      </template>
    </Header>
    <div class="places-body">
      <PlacesTable v-if="loggedIn" />
      <p v-else class="places-signin">Sign in to see places.</p>
    </div>

    <Modal v-model="showAdd">
      <CreatePlace @done="showAdd = false" @cancel="showAdd = false" />
    </Modal>
  </div>
</template>

<style scoped>
.places-cell {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.places-body {
  flex: 1;
  min-height: 0;
  position: relative;
  overflow: hidden;
}

.places-signin {
  padding: 12px;
  text-align: center;
  color: var(--color-muted);
}
</style>
