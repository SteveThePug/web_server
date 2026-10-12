/**
 * The places the lead fetcher polls, edited on /cv/leads — CRUD built on
 * createCrudStore().
 *
 * `lastFetchedAt`, `lastCount` and `lastError` are written by the fetcher on
 * the server, not by any mutation here, so the page calls fetch() again after
 * a run to see them move.
 */

import { createCrudStore } from "@/stores/createCrudStore";

const SOURCE_FIELDS = `id kind label token query location enabled lastFetchedAt lastCount lastError`;

export const useJobSourcesStore = createCrudStore("jobSources", {
  type: "JobSource",
  list: "jobSources",
  itemsKey: "sources",
  fields: SOURCE_FIELDS,
  // The list comes back ordered by label; a new row goes last until the next
  // fetch sorts it into place.
  insert: "append",
  // deleteJobSource returns Boolean!, and errors for an id that is gone.
  deleteReturns: "boolean",
});
