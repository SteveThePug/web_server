/**
 * Media/activity log for the Consumption widget. A view over homeData.activities,
 * built by createMirrorStore(); see stores/createMirrorStore.js for the
 * placeholder-then-overwrite pattern.
 */

import { defineStore } from "pinia";
import { createMirrorStore } from "@/stores/createMirrorStore";

const activity_template = {
  type: "activity",
  name: "nameof",
  createdAt: Date.now(),
};

export const useActivityStore = defineStore("activity", () => {
  const { items, count, fetch } = createMirrorStore({
    slice: "activities",
    template: activity_template,
  });

  return {
    activity: items,
    activityCount: count,
    fetchActivity: fetch,
  };
});
