import axios from "axios";

export async function gql(query, variables = {}) {
  const res = await axios.post("/api/graphql", { query, variables });
  if (res.data.errors && !res.data.data)
    throw new Error(res.data.errors[0].message);
  return res.data.data;
}
