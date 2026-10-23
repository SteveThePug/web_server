/**
 * Boxes on one canvas page (views/home3/BoxCanvas.vue) — full CRUD, built on
 * createCrudStore().
 *
 * `useBoxesStore(page)` returns the store for that page key: "index" for the
 * landing page, or a Page slug ("stp" for the home page). Each key
 * gets its own Pinia store (id `boxes:<page>`), so moving between pages
 * neither refetches a page already visited nor mixes two pages' boxes.
 *
 * Not part of the homeData query: the canvas has to know which boxes exist
 * before it can mount the widgets that then trigger homeData, and keeping the
 * two apart means a failing Spotify/Steam upstream can't blank the layout.
 *
 * The list query returns boxes ordered by z, so a new box (given the top z)
 * belongs at the end. adminOnly boxes only come back for an admin — the
 * server filters them — so the canvas refetches when the session changes.
 *
 * A box with a `componentId` is linked to that component (stores/components.js):
 * the server answers with the component's html in `html`, and editing it goes
 * to the component. patchLinked() keeps every page already loaded in step when
 * a component changes, without refetching them.
 */

import { createCrudStore } from "@/stores/createCrudStore";

const BOX_FIELDS = `
  id
  page
  kind
  widget
  html
  title
  x
  y
  w
  h
  z
  framed
  shadow
  hideOnMobile
  adminOnly
  componentId
`;

// The fields a duplicate copies. kind and widget/html are what make it the
// same box — or componentId, which keeps a copy of a linked box linked;
// geometry is offset so the copy is visibly separate.
const COPIED = [
  "page", "kind", "widget", "html", "title", "w", "h", "framed", "shadow", "hideOnMobile", "adminOnly", "componentId",
];

const stores = new Map();

function defineFor(page) {
  return createCrudStore(`boxes:${page}`, {
    type: "Box",
    list: "boxes",
    fields: BOX_FIELDS,
    insert: "append",
    deleteReturns: "object",
    args: { page: { type: "String!", value: page } },
    extend: ({ items, create, update }) => {
      const topZ = () => Math.max(0, ...items.value.map((b) => b.z)) + 1;
      const bottomZ = () => Math.min(0, ...items.value.map((b) => b.z)) - 1;

      return {
        page,
        topZ,
        // Every box this store creates belongs to its page.
        create: (input) => create({ ...input, page }),
        bringToFront: (box) => update(box.id, { z: topZ() }),
        sendToBack: (box) => update(box.id, { z: bottomZ() }),
        duplicate: (box) => {
          const input = Object.fromEntries(COPIED.map((k) => [k, box[k]]));
          return create({ ...input, x: box.x + 20, y: box.y + 20, z: topZ() });
        },
        /** Swap in a box the server answered with from outside this store. */
        replaceLocal: (box) => {
          const idx = items.value.findIndex((b) => b.id === box.id);
          if (idx >= 0) items.value[idx] = box;
        },
      };
    },
  });
}

/** The boxes store for one page key. */
export function useBoxesStore(page) {
  if (!stores.has(page)) stores.set(page, defineFor(page));
  return stores.get(page)();
}

/**
 * Apply `patch` to every loaded box linked to component `componentId`, on
 * every page visited so far.
 */
export function patchLinked(componentId, patch) {
  for (const use of stores.values()) {
    for (const box of use().boxes) {
      if (box.componentId === componentId) Object.assign(box, patch);
    }
  }
}
