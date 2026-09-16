// The shared body of the admin "create X" forms (CreatePost, CreateFavorite,
// CreateActivity, CreateBookmark): a ref per field, a GraphQL mutation taking a
// single `input`, clear the fields on success, then emit "done" so the host
// modal can close. Failures are logged and leave the fields alone, exactly as
// the hand-written versions did.

import { ref } from "vue";
import { gql } from "@/graphql";

/**
 * @param {object} options
 * @param {string} options.mutation   GraphQL mutation with one $input variable
 * @param {string[]} options.fields   input field names, one text ref each
 * @param {(event: string) => void} options.emit  the component's emit
 * @param {string[]} [options.optional]  fields sent as undefined when blank,
 *        so the backend sees "not set" rather than an empty string
 * @param {string} [options.resultKey]   mutation field to log on success
 * @returns {{ values: Record<string, import("vue").Ref<string>>, submit: () => Promise<void> }}
 */
export function useCreateForm({
  mutation,
  fields,
  emit,
  optional = [],
  resultKey,
}) {
  const values = Object.fromEntries(fields.map((name) => [name, ref("")]));

  async function submit() {
    try {
      const input = {};
      for (const name of fields) {
        const value = values[name].value;
        input[name] = optional.includes(name) ? value || undefined : value;
      }
      const data = await gql(mutation, { input });
      for (const name of fields) values[name].value = "";
      if (resultKey) console.log(data[resultKey]);
      emit("done");
    } catch (err) {
      console.error(err);
    }
  }

  return { values, submit };
}
