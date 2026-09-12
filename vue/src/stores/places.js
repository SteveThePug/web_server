import { defineStore } from "pinia";
import { ref } from "vue";
import { gql } from "@/graphql";

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

  async function toggleDone(place) {
    return update(place.id, { done: !place.done });
  }

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
