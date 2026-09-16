/**
 * The app's entire GraphQL client: a thin axios POST to the Go backend.
 * Every Pinia store goes through this; there is no cache or normalisation layer.
 *
 * Auth rides along in HTTP-only cookies set by the backend, so no token handling
 * is needed here — axios sends them automatically because the request is
 * same-origin (nginx in production, the Vite proxy in dev).
 */

import axios from "axios";

export async function gql(query, variables = {}) {
  const res = await axios.post("/api/graphql", { query, variables });
  if (res.data.errors && !res.data.data)
    throw new Error(res.data.errors[0].message);
  return res.data.data;
}
