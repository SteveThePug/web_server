/**
 * The REST half of the API client, beside graphql.js: a thin axios wrapper for
 * the Go backend's JSON endpoints under /api (file uploads, admin tooling,
 * OAuth starts — everything backend/README.md lists as "not GraphQL").
 *
 * It does two things a bare axios call does not:
 *   - prefixes `/api`, so callers name the route (`api.get("/radio/songs")`);
 *   - turns a failure into an ApiError whose `message` is the backend's
 *     `{"error": "..."}` body when there is one. Every handler reports errors
 *     in that shape, and each view used to dig it out of `err.response` itself.
 *
 * Resolves to the parsed response body, not the axios response.
 *
 * Auth is the same HTTP-only cookies graphql.js relies on; nothing here handles
 * tokens.
 *
 * NOT for: the Python service's NDJSON streams (views/hotels/hotelsStream.js
 * reads those incrementally with fetch) or the `/radio/stream` HEAD probe —
 * neither is under /api nor answers with this error shape.
 */

import axios from "axios";

export class ApiError extends Error {
  /**
   * @param {string} message
   * @param {number|undefined} status  HTTP status; undefined when no response
   *   arrived at all (network down, backend restarting)
   * @param {*} data  the parsed error body, if any
   */
  constructor(message, status, data) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.data = data;
  }
}

/**
 * @param {string} method
 * @param {string} path    route below /api, with its leading slash
 * @param {object} [options]
 * @param {*}      [options.body]      JSON-serialisable body, or a FormData for
 *   multipart uploads (axios sets the boundary itself)
 * @param {string} [options.fallback]  message to use when the backend sent no
 *   `error` string — callers pass the sentence their UI shows
 * Any other option (`headers`, `params`, `onUploadProgress`, `signal`, …) is
 * handed to axios unchanged.
 */
async function request(method, path, { body, fallback, ...config } = {}) {
  try {
    const res = await axios.request({
      method,
      url: `/api${path}`,
      data: body,
      ...config,
    });
    return res.data;
  } catch (err) {
    const data = err.response?.data;
    const serverMessage = typeof data?.error === "string" ? data.error : "";
    throw new ApiError(
      serverMessage || fallback || err.message,
      err.response?.status,
      data,
    );
  }
}

export const api = {
  get: (path, options) => request("get", path, options),
  delete: (path, options) => request("delete", path, options),
  post: (path, body, options) => request("post", path, { ...options, body }),
  put: (path, body, options) => request("put", path, { ...options, body }),
  patch: (path, body, options) => request("patch", path, { ...options, body }),
};
