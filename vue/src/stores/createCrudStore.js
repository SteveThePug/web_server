/**
 * Factory for the "CRUD store" pattern shared by stores/places.js,
 * jobApplications.js and jobAppReferences.js.
 *
 * Each of those is one GraphQL type with the schema's standard quartet — a list
 * query, `create<Type>(input:)`, `update<Type>(id:, input:)` and
 * `delete<Type>(id:)` — plus a local array that is patched after each mutation
 * instead of being refetched. The operation and input-type names are derived
 * from `type` because the schema is consistent about them
 * (`Create<Type>Input`, `Update<Type>Input`); a type that breaks the convention
 * should not be forced through here.
 *
 * `fields` is interpolated into every query and mutation so the selection set
 * can never drift between them.
 *
 * Error contract, which callers rely on: fetch() never throws — it records the
 * failure on `error` — because it is fired from lifecycle hooks with nothing to
 * catch it. create/update/remove DO throw, so the form or row that triggered
 * them can show the message and keep the user's typing.
 */

import { defineStore } from "pinia";
import { ref } from "vue";
import { gql } from "@/graphql";

/**
 * @param {string} id   Pinia store id
 * @param {object} options
 * @param {string} options.type     GraphQL type name, e.g. "Place"
 * @param {string} options.list     list query field, e.g. "places"
 * @param {string} options.fields   selection set shared by every operation
 * @param {string} [options.itemsKey]  name the array is exposed under (defaults
 *   to `list`); components import it by that name
 * @param {"prepend"|"append"} [options.insert]  where create() puts the new row
 *   locally, matching the order the list query returns
 * @param {"object"|"boolean"} [options.deleteReturns]  what the delete mutation
 *   returns. An object type needs a selection set and a scalar must not have
 *   one, so the mutation text differs; the schema has both kinds.
 * @param {(ctx: object) => object} [options.extend]  extra state/actions merged
 *   into the store; receives the base members
 */
export function createCrudStore(
  id,
  {
    type,
    list,
    fields,
    itemsKey = list,
    insert = "prepend",
    deleteReturns = "object",
    extend,
  },
) {
  const deleteSelection = deleteReturns === "object" ? "{ id }" : "";

  return defineStore(id, () => {
    const items = ref([]);
    const loaded = ref(false);
    const loading = ref(false);
    const error = ref(null);

    /** Load the whole list. Records failures on `error` rather than throwing. */
    async function fetch() {
      loading.value = true;
      try {
        const data = await gql(`query ${type}List { ${list} { ${fields} } }`);
        items.value = data[list] || [];
        loaded.value = true;
        error.value = null;
      } catch (err) {
        error.value = err;
      } finally {
        loading.value = false;
      }
    }

    /** Create a row and add it locally. Throws on failure. */
    async function create(input) {
      const data = await gql(
        `mutation Create${type}($input: Create${type}Input!) {
          create${type}(input: $input) { ${fields} }
        }`,
        { input },
      );
      const created = data[`create${type}`];
      items.value =
        insert === "append"
          ? [...items.value, created]
          : [created, ...items.value];
      return created;
    }

    /** Patch a row and replace it in the local list. Throws on failure. */
    async function update(rowId, input) {
      const data = await gql(
        `mutation Update${type}($id: ID!, $input: Update${type}Input!) {
          update${type}(id: $id, input: $input) { ${fields} }
        }`,
        { id: rowId, input },
      );
      const updated = data[`update${type}`];
      const idx = items.value.findIndex((item) => item.id === rowId);
      if (idx >= 0) items.value[idx] = updated;
      return updated;
    }

    /** Delete a row and drop it from the local list. Throws on failure. */
    async function remove(rowId) {
      await gql(
        `mutation Delete${type}($id: ID!) { delete${type}(id: $id) ${deleteSelection} }`,
        { id: rowId },
      );
      items.value = items.value.filter((item) => item.id !== rowId);
    }

    const base = { loaded, loading, error, fetch, create, update, remove };
    return {
      [itemsKey]: items,
      ...base,
      ...(extend ? extend({ items, ...base }) : {}),
    };
  });
}
