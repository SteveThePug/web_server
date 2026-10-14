/**
 * Places-to-go list (the "Places" tab inside the Collage widget) — full CRUD,
 * built on createCrudStore().
 *
 * Not part of the homeData query because it is sign-in-only; PlacesTable.vue
 * fetches it lazily once the user is logged in.
 */

import { createCrudStore } from "@/stores/createCrudStore";

const PLACE_FIELDS = `
  id
  title
  location
  notes
  imageUrl
  category
  cost
  priority
  done
  createdAt
  updatedAt
`;

export const usePlacesStore = createCrudStore("places", {
  type: "Place",
  list: "places",
  fields: PLACE_FIELDS,
  // The list query returns newest first, so a new place goes on top.
  insert: "prepend",
  // deletePlace returns the deleted Place rather than a Boolean.
  deleteReturns: "object",
  extend: ({ update }) => ({
    /** Flip a place's `done` flag. */
    toggleDone: (place) => update(place.id, { done: !place.done }),
  }),
});
