/**
 * Boxes on the /stp canvas (views/home3/) — full CRUD, built on
 * createCrudStore().
 *
 * Not part of the homeData query: the canvas has to know which boxes exist
 * before it can mount the widgets that then trigger homeData, and keeping the
 * two apart means a failing Spotify/Steam upstream can't blank the layout.
 *
 * The list query returns boxes ordered by z, so a new box (given the top z)
 * belongs at the end.
 */

import { createCrudStore } from "@/stores/createCrudStore";

const BOX_FIELDS = `
  id
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
  hideOnMobile
`;

// The fields a duplicate copies. kind and widget/html are what make it the
// same box; geometry is offset so the copy is visibly separate.
const COPIED = ["kind", "widget", "html", "title", "w", "h", "framed", "hideOnMobile"];

export const useBoxesStore = createCrudStore("boxes", {
  type: "Box",
  list: "boxes",
  fields: BOX_FIELDS,
  insert: "append",
  deleteReturns: "object",
  extend: ({ items, create, update }) => {
    const topZ = () => Math.max(0, ...items.value.map((b) => b.z)) + 1;
    const bottomZ = () => Math.min(0, ...items.value.map((b) => b.z)) - 1;

    return {
      topZ,
      bringToFront: (box) => update(box.id, { z: topZ() }),
      sendToBack: (box) => update(box.id, { z: bottomZ() }),
      duplicate: (box) => {
        const input = Object.fromEntries(COPIED.map((k) => [k, box[k]]));
        return create({ ...input, x: box.x + 20, y: box.y + 20, z: topZ() });
      },
    };
  },
});
