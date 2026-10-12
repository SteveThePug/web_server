/**
 * Job leads for the inbox at /cv/leads: postings the server's fetcher found
 * (backend/services/job_leads.go), which are dismissed or promoted into the
 * tracker from that page.
 *
 * Not a createCrudStore(): leads are never created or edited from the browser.
 * The mutations here move a lead between its three states, and each returns
 * the lead so the local row is replaced rather than the list refetched.
 *
 * Error contract matches the CRUD stores: fetch() and fetchStatus() record a
 * failure on `error` and never throw, because they run from lifecycle hooks
 * and a poll timer; everything else throws.
 */

import { defineStore } from "pinia";
import { ref } from "vue";
import { gql } from "@/graphql";

// `description` is deliberately not here: it is the full text of the posting
// and is fetched per lead when asked for, as on the tracker.
const LEAD_FIELDS = `id createdAt sourceKind jobTitle company location url postedAt lastSeenAt matched state score scoreReasons scoredAt jobAppId`;

export const useJobLeadsStore = defineStore("jobLeads", () => {
  const leads = ref([]);
  const loading = ref(false);
  const error = ref(null);
  // { running, lastRunAt, lastError } for the fetcher; null until first read.
  const status = ref(null);

  /**
   * Load every lead, all states. `includeUnmatched` also returns the ones the
   * keyword pre-filter rejected, which is most of a company board.
   */
  async function fetch({ includeUnmatched = false } = {}) {
    loading.value = true;
    try {
      const data = await gql(
        `query JobLeads($includeUnmatched: Boolean) {
          jobLeads(includeUnmatched: $includeUnmatched) { ${LEAD_FIELDS} }
        }`,
        { includeUnmatched },
      );
      leads.value = data.jobLeads || [];
      error.value = null;
    } catch (err) {
      error.value = err;
    } finally {
      loading.value = false;
    }
  }

  function replace(lead) {
    const idx = leads.value.findIndex((l) => l.id === lead.id);
    if (idx >= 0) leads.value[idx] = lead;
    return lead;
  }

  async function transition(mutation, id) {
    const data = await gql(
      `mutation ${mutation}($id: ID!) { ${mutation}(id: $id) { ${LEAD_FIELDS} } }`,
      { id },
    );
    return replace(data[mutation]);
  }

  /** Each returns the updated lead and throws on failure. */
  const dismiss = (id) => transition("dismissJobLead", id);
  const restore = (id) => transition("restoreJobLead", id);
  /** Adds the lead to the tracker as a saved application. */
  const promote = (id) => transition("promoteJobLead", id);

  /** The posting text for one lead; `""` when the source gave none. Throws. */
  async function fetchDescription(id) {
    const data = await gql(
      `query JobLeadDescription($id: ID!) { jobLead(id: $id) { description } }`,
      { id },
    );
    return data.jobLead?.description ?? "";
  }

  async function fetchStatus() {
    try {
      const data = await gql(`query JobLeadSyncStatus { jobLeadSyncStatus { running lastRunAt lastError } }`);
      status.value = data.jobLeadSyncStatus;
    } catch (err) {
      error.value = err;
    }
    return status.value;
  }

  /**
   * Start a fetch-and-score run on the server. Resolves at once — true if it
   * started, false if one was already going — and the caller polls
   * fetchStatus() to see it finish. Throws on failure.
   */
  async function runSync() {
    const data = await gql(`mutation RunJobLeadSync { runJobLeadSync }`);
    return data.runJobLeadSync;
  }

  /**
   * Score every unscored lead in the inbox without polling any source. Same
   * contract as runSync(): resolves at once, and fetchStatus() shows it finish.
   */
  async function scoreAll() {
    const data = await gql(`mutation ScoreJobLeads { scoreJobLeads }`);
    return data.scoreJobLeads;
  }

  /**
   * Score every inbox lead again, scored or not — for after the profile or
   * the model has changed. Same contract as runSync(). Each lead keeps its old
   * score until its new one lands, so the list never goes blank.
   */
  async function rescoreAll() {
    const data = await gql(`mutation RescoreJobLeads { rescoreJobLeads }`);
    return data.rescoreJobLeads;
  }

  return { leads, loading, error, status, fetch, dismiss, restore, promote, fetchDescription, fetchStatus, runSync, scoreAll, rescoreAll };
});
