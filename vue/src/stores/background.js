/**
 * Page backgrounds (the `Background` GraphQL type), one per canvas page key —
 * "index" for the landing page, "stp", or a Page slug such as
 * "shrines/evangelion" — and which one is on screen.
 *
 * DefaultLayout paints `style` behind everything, navbar included, so a
 * background covers the whole window rather than just the canvas. Each
 * BoxCanvas calls show(page) when it mounts and hide(page) when it goes;
 * a page with no background (null) keeps the theme's halftone.
 *
 * `preview` is BackgroundEditor's unsaved draft: while it is set for the page
 * on screen, it is what the layout paints, so every change shows at once.
 */

import { defineStore } from "pinia";
import { computed, ref } from "vue";
import { gql } from "@/graphql";

const FIELDS = "id page color image fit";

export const FITS = [
  { id: "cover", label: "Fill the window" },
  { id: "contain", label: "Fit inside the window" },
  { id: "tile", label: "Tile" },
];

/** CSS for a background, or null for the theme default. */
export function backgroundStyle(bg) {
  if (!bg || (!bg.color && !bg.image)) return null;
  const style = {
    backgroundColor: bg.color || "var(--color-surface-deep)",
    // "none" as well as a URL: either way the halftone's own layers go.
    backgroundImage: bg.image ? `url("${bg.image}")` : "none",
  };
  if (bg.fit === "tile") {
    Object.assign(style, { backgroundSize: "auto", backgroundRepeat: "repeat", backgroundPosition: "0 0" });
  } else {
    // Pinned to the window, so a tall page scrolls over one picture rather
    // than stretching it down the whole page.
    Object.assign(style, {
      backgroundSize: bg.fit,
      backgroundRepeat: "no-repeat",
      backgroundPosition: "center",
      backgroundAttachment: "fixed",
    });
  }
  return style;
}

export const useBackgroundStore = defineStore("background", () => {
  // page key -> Background or null; a key not present hasn't been loaded.
  const byPage = ref({});
  const active = ref(null);
  const preview = ref(null); // { page, color, image, fit }

  const current = computed(() =>
    preview.value && preview.value.page === active.value ? preview.value : byPage.value[active.value],
  );
  const style = computed(() => backgroundStyle(current.value));

  async function load(page) {
    try {
      const data = await gql(`query Background($page: String!) { background(page: $page) { ${FIELDS} } }`, {
        page,
      });
      byPage.value[page] = data.background;
    } catch (err) {
      // The theme default is a fine fallback; the page itself still loads.
      console.warn(`Couldn't load the background for ${page}:`, err);
    }
  }

  function show(page) {
    active.value = page;
    if (!(page in byPage.value)) load(page);
  }

  function hide(page) {
    if (active.value === page) active.value = null;
    if (preview.value?.page === page) preview.value = null;
  }

  async function save(page, { color, image, fit }) {
    const data = await gql(
      `mutation SetBackground($page: String!, $input: BackgroundInput!) {
        setBackground(page: $page, input: $input) { ${FIELDS} }
      }`,
      { page, input: { color, image, fit } },
    );
    byPage.value[page] = data.setBackground;
    return data.setBackground;
  }

  async function clear(page) {
    await gql(`mutation ClearBackground($page: String!) { clearBackground(page: $page) }`, { page });
    byPage.value[page] = null;
  }

  return { byPage, active, preview, current, style, load, show, hide, save, clear };
});
