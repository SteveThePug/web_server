/**
 * The cover-letter paragraph bank (the `CoverSnippet` GraphQL type): reusable
 * paragraphs dragged into a letter as copies — full CRUD, built on
 * createCrudStore().
 *
 * `section` is the title of the section a snippet was written for and only
 * groups the bank (views/CV/CoverSnippetBank.vue); editing a snippet never
 * changes letters it was already dropped into.
 *
 * Admin only, like stores/coverLetters.js.
 */

import { createCrudStore } from "@/stores/createCrudStore";

const COVER_SNIPPET_FIELDS = `
  id
  text
  section
  createdAt
`;

export const useCoverSnippetsStore = createCrudStore("coverSnippets", {
  type: "CoverSnippet",
  list: "coverSnippets",
  fields: COVER_SNIPPET_FIELDS,
  // The bank groups by section itself, so where a new one lands doesn't matter.
  insert: "prepend",
  deleteReturns: "object",
});
