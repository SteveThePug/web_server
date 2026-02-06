import { defineStore } from "pinia";
import { computed, ref } from "vue";
import axios from "axios";

const activity_template = {
  type: "activity",
  name: "nameof",
  createdAt: Date.now(),
};

export const useActivityStore = defineStore("activity", () => {
  const activity = ref([activity_template]);

  const activityCount = computed(() => activity.value.length);

  async function fetchActivity() {
    try {
      const res = await axios.get("/api/activity");
      if (!Array.isArray(res.data)) {
        throw new Error("Invalid response from posts API");
      }
      activity.value = res.data;
    } catch (err) {
      console.error("Cannot connect to activity API", err);
    }
  }

  return {
    activity,
    activityCount,
    fetchActivity,
  };
});
