/**
 * The single search profile on /cv/leads: the free text every lead is scored
 * against, the keyword lists behind the pre-filter, and the inbox's minimum
 * score. One row on the server, so this is a fetch and a save rather than a
 * createCrudStore() list.
 *
 * Error contract matches the CRUD stores: fetch() records its failure on
 * `error` and never throws; save() throws so the form can keep the typing.
 */

import { defineStore } from "pinia";
import { ref } from "vue";
import { gql } from "@/graphql";

const PROFILE_FIELDS = `profile includeKeywords excludeKeywords locations minScore updatedAt`;

export const useJobSearchProfileStore = defineStore("jobSearchProfile", () => {
  const profile = ref(null);
  const error = ref(null);

  async function fetch() {
    try {
      const data = await gql(`query JobSearchProfile { jobSearchProfile { ${PROFILE_FIELDS} } }`);
      profile.value = data.jobSearchProfile;
      error.value = null;
    } catch (err) {
      error.value = err;
    }
  }

  /** Save the edited fields. Throws on failure. */
  async function save(input) {
    const data = await gql(
      `mutation UpdateJobSearchProfile($input: UpdateJobSearchProfileInput!) {
        updateJobSearchProfile(input: $input) { ${PROFILE_FIELDS} }
      }`,
      { input },
    );
    profile.value = data.updateJobSearchProfile;
    return profile.value;
  }

  return { profile, error, fetch, save };
});
