/**
 * The job-application status vocabulary, as the tracker at /cv/jobs shows it.
 *
 * This is a COPY. The source of truth is server-side: `enum JobStatus` in
 * backend/graph/schema/job_application.graphql, ranked by `statusOrder` in
 * backend/services/email_sync.go (the email pipeline only ever moves a row
 * forward through that order). The server rejects any value outside the enum,
 * so a status added here alone cannot be saved — add it there first.
 *
 * There is a third copy in vue/public/job-capture.user.js (STATUS_LABELS). A
 * userscript is a single file installed into the browser and cannot import
 * modules, so it has to be kept in step by hand.
 *
 * Values are the lowercase keys that go over the wire; labels are display-only.
 */

export const STATUS_OPTIONS = [
  // Captured from the browser but not applied to yet. No email produces it.
  { value: "saved", label: "Saved" },
  { value: "applied", label: "Applied" },
  { value: "screening", label: "Screening" },
  { value: "assessment", label: "Assessment" },
  { value: "interviewing", label: "Interviewing" },
  { value: "offer", label: "Offer" },
  { value: "rejected", label: "Rejected" },
  { value: "withdrawn", label: "Withdrawn" },
];

// Statuses still in play — used to decide whether a row is worth chasing.
// "saved" is absent on purpose: nobody owes a reply on a job not applied to.
export const OPEN_STATUSES = ["applied", "screening", "assessment", "interviewing"];
