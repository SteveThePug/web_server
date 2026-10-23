/**
 * Saved layouts (the `Layout` GraphQL type): named copies of a canvas page's
 * boxes and background — full CRUD, built on createCrudStore(), plus apply().
 *
 * `create({ name, page })` saves that page as it is now; `update` only
 * renames. What a layout holds never reaches the browser, only its size: the
 * server does both the copying in and the copying out.
 *
 * Admin only, so nothing fetches it until an admin opens a layouts list
 * (views/home3/LayoutPanel.vue, the admin panel's Pages tab).
 */

import { createCrudStore } from "@/stores/createCrudStore";
import { gql } from "@/graphql";
import { useBoxesStore } from "@/stores/boxes";
import { useBackgroundStore } from "@/stores/background";

const LAYOUT_FIELDS = `
  id
  name
  sourcePage
  boxCount
  hasBackground
  createdAt
`;

export const useLayoutsStore = createCrudStore("layouts", {
  type: "Layout",
  list: "layouts",
  fields: LAYOUT_FIELDS,
  // The list query is newest first.
  insert: "prepend",
  deleteReturns: "object",
  extend: ({ fetch }) => ({
    /**
     * Copy layout `id` onto page key `page`, replacing what is there or
     * adding on top. Throws on failure.
     */
    async apply(id, page, replace) {
      await gql(
        `mutation ApplyLayout($id: ID!, $page: String!, $replace: Boolean!) {
          applyLayout(id: $id, page: $page, replace: $replace) { id }
        }`,
        { id, page, replace },
      );
      // The page's own stores are refetched rather than patched from the
      // answer, so they stay in the shape their list query gives. A replace
      // also saved the old page as a new layout, hence the list too.
      await Promise.all([useBoxesStore(page).fetch(), useBackgroundStore().load(page), replace && fetch()]);
    },
  }),
});

/** How a page key reads to a person: "/" for the landing page. */
export function pagePath(page) {
  return page === "index" ? "/" : `/${page}`;
}
