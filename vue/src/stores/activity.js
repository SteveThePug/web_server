/**
 * Media/activity log for the Consumption widget. A view over homeData.activities;
 * see stores/posts.js for the placeholder-then-overwrite pattern.
 */

import { defineStore } from "pinia";
import { computed, ref, watch } from "vue";
import { useHomeDataStore } from "@/stores/homeData";

const activity_template = {
  type: "activity",
  name: "nameof",
  createdAt: Date.now(),
};

export const useActivityStore = defineStore("activity", () => {
  const activity = ref([activity_template]);

  const activityCount = computed(() => activity.value.length);

  const homeData = useHomeDataStore();
  // Mirror the shared home query. `immediate` matters: homeData may already
  // have loaded by the time this store is first used, and a plain watch would
  // never fire for that existing value. The `length > 0` test keeps the
  // placeholder on screen when the backend returns nothing.
  watch(
    () => homeData.activities,
    (newActivities) => {
      if (newActivities.length > 0) {
        activity.value = newActivities;
      }
    },
    { immediate: true },
  );

  /** Refresh by re-running the whole home query; the watch above applies it. */
  async function fetchActivity() {
    await homeData.fetchAll();
  }

  return {
    activity,
    activityCount,
    fetchActivity,
  };
});
