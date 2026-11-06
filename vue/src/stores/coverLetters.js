/**
 * Saved cover letters (the `CoverLetter` GraphQL type) for the /cv Cover
 * Letter tab — full CRUD, built on createCrudStore().
 *
 * `body` is the whole letter as a JSON string; views/CV/coverLetter.js owns
 * its shape (parseBody / serialize), the server only checks it is an object.
 *
 * Admin only, so nothing fetches it until views/CV/CoverLetters.vue sees an
 * admin signed in.
 */

import { createCrudStore } from "@/stores/createCrudStore";

const COVER_LETTER_FIELDS = `
  id
  name
  body
  updatedAt
`;

export const useCoverLettersStore = createCrudStore("coverLetters", {
  type: "CoverLetter",
  list: "coverLetters",
  fields: COVER_LETTER_FIELDS,
  // The list query is most recently updated first, so a new letter goes on top.
  insert: "prepend",
  deleteReturns: "object",
});
