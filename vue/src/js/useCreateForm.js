// The shared body of the admin "create X" forms (CreatePost, CreateFavorite,
// CreateActivity, CreateBookmark): a ref per field, a GraphQL mutation taking a
// single `input`, clear the fields on success, then emit "done" so the host
// modal can close. A failure leaves the fields alone and its message on
// `error` for the form to show — it used to go only to the console, so a
// refused save looked exactly like a button that did nothing. `submitting`
// disables the button while the request is out, and a second submit in that
// window is ignored, so a double click cannot create the row twice.

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
 * @returns {{
 *   values: Record<string, import("vue").Ref<string>>,
 *   submit: () => Promise<void>,
 *   error: import("vue").Ref<string>,
 *   submitting: import("vue").Ref<boolean>,
 * }}
 */
export function useCreateForm({
  mutation,
  fields,
  emit,
  optional = [],
  resultKey,
}) {
  const values = Object.fromEntries(fields.map((name) => [name, ref("")]));
  const error = ref("");
  const submitting = ref(false);

  async function submit() {
    if (submitting.value) return;
    submitting.value = true;
    error.value = "";
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
      error.value = err?.message || "Failed to save";
    } finally {
      submitting.value = false;
    }
  }

  return { values, submit, error, submitting };
}
