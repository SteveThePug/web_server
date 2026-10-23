/**
 * Current user + login/logout mutations.
 *
 * `user` is populated three ways:
 *   - ensureLoaded(): a standalone `me` query, fired at most once per page load.
 *     The router guard awaits it, and so does the login page.
 *   - adoptSession(): stores/homeData.js hands over the `me` it got back from
 *     the home query, so the /stp page never needs the standalone query.
 *   - logIn() / refreshToken() / logOut(), which set it directly.
 *
 * This store deliberately does NOT import homeData. It used to mirror
 * homeData.me with a watch, which meant that merely asking "who is signed in?"
 * instantiated homeData and so fired the nine-field home query — Spotify, Steam
 * and Gitea upstreams included — and the guard then had to wait on it. The
 * dependency now points the other way: homeData pushes `me` in here.
 *
 * The JWTs themselves are never visible here — the backend sets access_token and
 * refresh_token as HTTP-only cookies, so `loggedIn` is inferred purely from
 * whether a username came back.
 */

import { defineStore } from "pinia";
import { computed, ref } from "vue";
import { gql } from "@/graphql";

// The user fields every query here selects, so `user` has the same shape
// whichever one filled it. `permissions` is the list of gated services
// (/receipt/ and so on) a non-admin may open; see views/Account.vue.
export const USER_FIELDS = "id username admin permissions";

export const useAuthStore = defineStore("auth", () => {
  const user = ref({});
  const loggedIn = computed(() => !!user.value.username);

  // True once something has answered "who is signed in?" — the `me` query,
  // the home query, or a login/logout. Until then an empty `user` means
  // "unknown", not "signed out", which is the distinction the guard needs.
  const resolved = ref(false);

  // Bumped by every login/logout/refresh. A `me` answer that was requested
  // before the bump describes the previous session, so it is dropped rather
  // than allowed to sign a logged-out user back in (or the reverse).
  let epoch = 0;
  let pending = null;

  function markSessionChanged() {
    epoch++;
    resolved.value = true;
  }

  /**
   * Apply a `me` answer that was requested when the session was at
   * `requestedAt` (a value from sessionEpoch()). Only a truthy `me` overwrites
   * `user`: a signed-out answer is null and must not clobber a user that
   * logIn() set while the request was in flight.
   */
  function adoptSession(me, requestedAt = epoch) {
    if (requestedAt !== epoch) return;
    if (me) user.value = me;
    resolved.value = true;
  }

  /** Token for adoptSession(): read it before issuing a query that returns `me`. */
  function sessionEpoch() {
    return epoch;
  }

  /**
   * Resolve once the signed-in user is known, fetching `me` if nothing has
   * supplied it yet. NEVER rejects and never hangs: a failed query settles as
   * "not signed in", because the router guard awaits this and a navigation
   * that never resolves is worse than a bounce to the login page. A failure is
   * not remembered, so the next call tries again.
   */
  function ensureLoaded() {
    if (resolved.value) return Promise.resolve();
    if (!pending) {
      const requestedAt = epoch;
      pending = gql(`query Me { me { ${USER_FIELDS} } }`)
        .then((data) => adoptSession(data.me, requestedAt))
        .catch((err) => console.error("me query failed:", err))
        .finally(() => {
          pending = null;
        });
    }
    return pending;
  }

  /** Clear the session cookies server-side, then blank the local user.
   *  Deliberately clears `user` even if the mutation fails. */
  async function logOut() {
    try {
      await gql(`mutation { logout }`);
    } catch (err) {
      console.error(err);
    }
    user.value = {};
    markSessionChanged();
  }

  /** Log in and populate `user`. Swallows errors: on a bad password `user`
   *  is left untouched, so `loggedIn` simply stays false. */
  async function logIn(username, password) {
    try {
      const data = await gql(
        `mutation Login($input: LoginInput!) { login(input: $input) { user { ${USER_FIELDS} } } }`,
        { input: { username, password } },
      );
      user.value = data.login.user;
      markSessionChanged();
    } catch (err) {
      console.error(err);
    }
  }

  /** Admin-only: create an account. Returns the new user; re-throws on failure
   *  so the calling form can show a message. */
  async function createUser(username, password) {
    try {
      const data = await gql(
        `mutation CreateUser($input: CreateUserInput!) { createUser(input: $input) { ${USER_FIELDS} } }`,
        { input: { username, password } },
      );
      return data.createUser;
    } catch (err) {
      console.error(err);
      throw err;
    }
  }

  /** Exchange the long-lived refresh_token cookie for a fresh access_token and
   *  re-read the user. */
  async function refreshToken() {
    try {
      const data = await gql(
        `mutation { refreshToken { user { ${USER_FIELDS} } } }`,
      );
      user.value = data.refreshToken.user;
      markSessionChanged();
    } catch (err) {
      console.log(err);
    }
  }

  /** Admin-only: grant/revoke admin. Returns the updated user; re-throws. */
  async function setUserAdmin(userId, admin) {
    try {
      const data = await gql(
        `mutation SetUserAdmin($id: ID!, $admin: Boolean!) { setUserAdmin(id: $id, admin: $admin) { ${USER_FIELDS} } }`,
        { id: userId, admin },
      );
      return data.setUserAdmin;
    } catch (err) {
      console.error(err);
      throw err;
    }
  }

  /** Admin-only: replace a user's permission list. Returns the updated user;
   *  re-throws. */
  async function setUserPermissions(userId, permissions) {
    try {
      const data = await gql(
        `mutation SetUserPermissions($id: ID!, $permissions: [String!]!) { setUserPermissions(id: $id, permissions: $permissions) { ${USER_FIELDS} } }`,
        { id: userId, permissions },
      );
      return data.setUserPermissions;
    } catch (err) {
      console.error(err);
      throw err;
    }
  }

  /** Signed-in only: every grantable service, as [{ key, label, path }].
   *  Re-throws. */
  async function fetchPermissionTypes() {
    const data = await gql(`query { permissionTypes { key label path } }`);
    return data.permissionTypes;
  }

  /** Whether the signed-in user may open the service gated on `key`. Admins
   *  may open them all, as on the server. */
  function can(key) {
    return !!user.value.admin || (user.value.permissions ?? []).includes(key);
  }

  return {
    user,

    loggedIn,

    ensureLoaded,
    adoptSession,
    sessionEpoch,

    logIn,
    refreshToken,
    logOut,
    createUser,
    setUserAdmin,
    setUserPermissions,
    fetchPermissionTypes,
    can,
  };
});
