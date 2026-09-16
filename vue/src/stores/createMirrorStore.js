/**
 * Factory for the "mirror store" pattern shared by stores/posts.js,
 * favorites.js and activity.js.
 *
 * Each of those is a thin view over one slice of the shared home query
 * (stores/homeData.js): a ref seeded with placeholder content, an `immediate`
 * watch that copies the homeData slice in, a count computed and a fetch
 * function that just re-runs the whole home query.
 *
 * The state/getter/action names differ per store and components import them by
 * name, so the caller supplies them explicitly rather than having them derived.
 */

import { computed, ref, watch } from "vue";
import { useHomeDataStore } from "@/stores/homeData";

/**
 * @param {object} options
 * @param {string} options.slice     key on the homeData store to mirror
 * @param {*}      options.template  placeholder item shown before a successful fetch
 * @returns {{ items: import("vue").Ref, count: import("vue").ComputedRef, fetch: () => Promise<void>, homeData: object }}
 */
export function createMirrorStore({ slice, template }) {
  const items = ref([template]);

  const count = computed(() => items.value.length);

  const homeData = useHomeDataStore();
  // `immediate` matters: homeData may already have loaded by the time this
  // store is first used, and a plain watch would never fire for that existing
  // value. The `length > 0` test keeps the placeholder on screen when the
  // backend returns nothing.
  watch(
    () => homeData[slice],
    (newItems) => {
      if (newItems.length > 0) {
        items.value = newItems;
      }
    },
    { immediate: true },
  );

  /** Refresh by re-running the whole home query; the watch above applies it. */
  async function fetch() {
    await homeData.fetchAll();
  }

  return { items, count, fetch, homeData };
}
