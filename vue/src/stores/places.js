/**
 * Places-to-go list (the "Places" tab inside the Collage widget) — full CRUD.
 *
 * Not part of the homeData query because it is sign-in-only; PlacesTable.vue
 * fetches it lazily once the user is logged in. PLACE_FIELDS is interpolated into
 * every query/mutation so the selection set can never drift between them.
 */

import { defineStore } from "pinia";
import { ref } from "vue";
import { gql } from "@/graphql";

// Shared selection set: interpolated into every query and mutation so they can
// never drift apart.
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

export const usePlacesStore = defineStore("places", () => {
  const places = ref([]);
  const loaded = ref(false);
  const loading = ref(false);
  const error = ref(null);

  /** Load all places into `places`. Records failures on `error` rather than throwing. */
  async function fetch() {
    loading.value = true;
    try {
      const data = await gql(`query Places { places { ${PLACE_FIELDS} } }`);
      places.value = data.places || [];
      loaded.value = true;
      error.value = null;
    } catch (err) {
      error.value = err;
    } finally {
      loading.value = false;
    }
  }

  /** Create a place and prepend it locally (newest first). Throws on failure. */
  async function create(input) {
    const data = await gql(
      `mutation CreatePlace($input: CreatePlaceInput!) {
        createPlace(input: $input) { ${PLACE_FIELDS} }
      }`,
      { input },
    );
    places.value = [data.createPlace, ...places.value];
    return data.createPlace;
  }

  /** Patch a place and replace it in the local list. Throws on failure. */
  async function update(id, input) {
    const data = await gql(
      `mutation UpdatePlace($id: ID!, $input: UpdatePlaceInput!) {
        updatePlace(id: $id, input: $input) { ${PLACE_FIELDS} }
      }`,
      { id, input },
    );
    const idx = places.value.findIndex((p) => p.id === id);
    if (idx >= 0) places.value[idx] = data.updatePlace;
    return data.updatePlace;
  }

  /** Flip a place's `done` flag. */
  async function toggleDone(place) {
    return update(place.id, { done: !place.done });
  }

  /** Soft-delete a place and drop it from the local list. */
  async function remove(id) {
    await gql(
      `mutation DeletePlace($id: ID!) { deletePlace(id: $id) { id } }`,
      { id },
    );
    places.value = places.value.filter((p) => p.id !== id);
  }

  return {
    places,
    loaded,
    loading,
    error,
    fetch,
    create,
    update,
    toggleDone,
    remove,
  };
});
