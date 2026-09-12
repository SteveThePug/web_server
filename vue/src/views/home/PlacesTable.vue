<script setup>
import { computed, onMounted } from "vue";
import { usePlacesStore } from "@/stores/places";
import { useAuthStore } from "@/stores/auth";

const places = usePlacesStore();
const auth = useAuthStore();
const isAdmin = computed(() => !!auth.user?.admin);

onMounted(() => {
  if (!places.loaded && !places.loading) places.fetch();
});

async function toggleDone(place) {
  if (!isAdmin.value) return;
  try {
    await places.toggleDone(place);
  } catch (err) {
    console.error(err);
  }
}

async function remove(place) {
  if (!isAdmin.value) return;
  if (!confirm(`Delete "${place.title}"?`)) return;
  try {
    await places.remove(place.id);
  } catch (err) {
    console.error(err);
  }
}
</script>

<template>
  <div class="places-table">
    <p v-if="!places.loaded && places.loading" class="places-empty">
      Loading…
    </p>
    <p v-else-if="places.error" class="places-empty">
      Couldn't load places.
    </p>
    <p v-else-if="!places.places.length" class="places-empty">
      Nothing here yet — add somewhere to go.
    </p>
    <table v-else class="places">
      <thead>
        <tr>
          <th>Title</th>
          <th class="col-loc">Where</th>
          <th class="col-cat">Type</th>
          <th class="col-cost">Cost</th>
          <th v-if="isAdmin" class="col-act"></th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="place in places.places"
          :key="place.id"
          :class="{ 'is-done': place.done }"
        >
          <td>
            <div class="title-cell">
              <button
                v-if="isAdmin"
                class="done-toggle"
                :title="place.done ? 'Mark as pending' : 'Mark as done'"
                @click="toggleDone(place)"
              >
                {{ place.done ? "✓" : "○" }}
              </button>
              <span class="title-text">{{ place.title }}</span>
            </div>
            <small v-if="place.notes" class="notes">{{ place.notes }}</small>
          </td>
          <td class="col-loc">{{ place.location || "—" }}</td>
          <td class="col-cat">{{ place.category || "—" }}</td>
          <td class="col-cost">{{ place.cost || "—" }}</td>
          <td v-if="isAdmin" class="col-act">
            <button class="row-action" @click="remove(place)" title="Delete">
              ×
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.places-table {
  width: 100%;
  height: 100%;
  overflow-y: auto;
}

.places-empty {
  padding: 12px;
  text-align: center;
  color: var(--color-muted);
}

table.places {
  width: 100%;
  border: none;
  border-collapse: collapse;
  font-size: 0.85rem;
}

table.places th {
  position: sticky;
  top: 0;
  background-color: var(--color-surface);
  border-right: 1px dotted var(--color-tertiary);
  border-bottom: 1px solid var(--color-primary);
  padding: 4px 6px;
  text-align: left;
  font-family: var(--font-heading);
}

table.places td {
  padding: 4px 6px;
  vertical-align: top;
  border-bottom: 1px dotted var(--color-quaternary);
}

.title-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.title-text {
  color: var(--color-primary);
  font-family: var(--font-heading);
}

.notes {
  display: block;
  color: var(--color-muted);
  font-size: 0.75rem;
  padding-top: 2px;
}

tr.is-done .title-text {
  text-decoration: line-through;
  color: var(--color-muted);
}

.done-toggle,
.row-action {
  background-color: transparent;
  color: var(--color-primary);
  border: 1px solid var(--color-quaternary);
  width: 20px;
  height: 20px;
  line-height: 1;
  cursor: pointer;
  font-family: var(--font-heading);
  padding: 0;
}
.done-toggle:hover,
.row-action:hover {
  border-color: var(--color-tertiary);
  color: var(--color-tertiary);
}

.col-cost,
.col-cat {
  white-space: nowrap;
}

@media (max-width: 700px) {
  .col-cost,
  .col-cat {
    display: none;
  }
}
</style>
