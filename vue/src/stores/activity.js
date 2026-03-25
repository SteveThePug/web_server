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
  watch(
    () => homeData.activities,
    (newActivities) => {
      if (newActivities.length > 0) {
        activity.value = newActivities;
      }
    },
    { immediate: true },
  );

  async function fetchActivity() {
    await homeData.fetchAll();
  }

  return {
    activity,
    activityCount,
    fetchActivity,
  };
});
