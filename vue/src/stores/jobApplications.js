/**
 * Job applications for the tracker at /cv/jobs — CRUD built on
 * createCrudStore(), plus the on-demand posting text.
 *
 * Admin-only on the server, and only that one page uses it, so nothing fetches
 * it until the view mounts.
 */

import { createCrudStore } from "@/stores/createCrudStore";
import { gql } from "@/graphql";

// `description` is deliberately not here: it is the full text of the posting,
// tens of kilobytes a row, and is fetched per row when asked for.
const APP_FIELDS = `id jobTitle company location url status notes appliedAt createdAt updatedAt`;

export const useJobApplicationsStore = createCrudStore("jobApplications", {
  type: "JobApplication",
  list: "jobApplications",
  itemsKey: "applications",
  fields: APP_FIELDS,
  insert: "prepend",
  // deleteJobApplication returns Boolean!, and errors for an id that is gone.
  deleteReturns: "boolean",
  extend: () => ({
    /**
     * The captured posting text for one row; `""` when none was saved. Throws
     * on failure. Not cached here: the view keeps its own per-visit cache, so
     * returning to the page re-reads a posting the userscript may have
     * replaced.
     */
    async fetchDescription(id) {
      const data = await gql(
        `query JobDescription($id: ID!) { jobApplication(id: $id) { description } }`,
        { id },
      );
      return data.jobApplication?.description ?? "";
    },
  }),
});
