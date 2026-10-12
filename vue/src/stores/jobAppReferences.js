/**
 * The "Quick Reference" snippets on /cv/jobs (reusable profile/experience text
 * to paste into application forms) — CRUD built on createCrudStore().
 *
 * A separate table from the applications themselves; the two only share a page.
 */

import { createCrudStore } from "@/stores/createCrudStore";

const REF_FIELDS = `id category label value sortOrder createdAt`;

export const useJobAppReferencesStore = createCrudStore("jobAppReferences", {
  type: "JobAppReference",
  list: "jobAppReferences",
  itemsKey: "references",
  fields: REF_FIELDS,
  // The list comes back in sortOrder/creation order, so new rows go last.
  insert: "append",
  // deleteJobAppReference returns Boolean!, and errors for an id that is gone.
  deleteReturns: "boolean",
});
