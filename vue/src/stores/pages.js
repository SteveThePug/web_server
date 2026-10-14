/**
 * Admin-created canvas pages (the `Page` GraphQL type) — full CRUD, built on
 * createCrudStore().
 *
 * A page is only a slug and a title; what is on it lives in that slug's boxes
 * store (stores/boxes.js). The list is public apart from adminOnly pages,
 * which the server only returns to an admin. Single pages are looked up by
 * views/canvas/CanvasPage.vue with its own `page(slug:)` query, so that
 * opening one page doesn't download the whole list.
 */

import { createCrudStore } from "@/stores/createCrudStore";

export const PAGE_FIELDS = `
  id
  slug
  title
  adminOnly
`;

export const usePagesStore = createCrudStore("pages", {
  type: "Page",
  list: "pages",
  fields: PAGE_FIELDS,
  insert: "append",
  deleteReturns: "object",
});
