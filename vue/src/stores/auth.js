/**
 * Current user + login/logout mutations.
 *
 * `user` is populated two ways: mirrored from homeData.me (the shared home query
 * also returns the session user), and set directly by logIn()/refreshToken().
 * The watch is `immediate` so a store created after homeData has already loaded
 * still picks up the user.
 *
 * The JWTs themselves are never visible here — the backend sets access_token and
 * refresh_token as HTTP-only cookies, so `loggedIn` is inferred purely from
 * whether a username came back.
 */

import { defineStore } from "pinia";
import { computed, ref, watch } from "vue";
import { gql } from "@/graphql";
import { useHomeDataStore } from "@/stores/homeData";

export const useAuthStore = defineStore("auth", () => {
  const user = ref({});
  const loggedIn = computed(() => !!user.value.username);

  const homeData = useHomeDataStore();
  // Mirror the `me` field from the shared home query. `immediate` so a store
  // created after homeData loaded still sees the user. Only overwrite on a
  // truthy `me` — a logged-out response is null and must not clobber a user
  // that logIn() just set.
  watch(
    () => homeData.me,
    (me) => {
      if (me) {
        user.value = me;
      }
    },
    { immediate: true },
  );

  /** Clear the session cookies server-side, then blank the local user.
   *  Deliberately clears `user` even if the mutation fails. */
  async function logOut() {
    try {
      await gql(`mutation { logout }`);
    } catch (err) {
      console.error(err);
    }
    user.value = {};
  }

  /** Log in and populate `user`. Swallows errors: on a bad password `user`
   *  is left untouched, so `loggedIn` simply stays false. */
  async function logIn(username, password) {
    try {
      const data = await gql(
        `mutation Login($input: LoginInput!) { login(input: $input) { user { id username admin } } }`,
        { input: { username, password } },
      );
      user.value = data.login.user;
    } catch (err) {
      console.error(err);
    }
  }

  /** Admin-only: create an account. Returns the new user; re-throws on failure
   *  so the calling form can show a message. */
  async function createUser(username, password) {
    try {
      const data = await gql(
        `mutation CreateUser($input: CreateUserInput!) { createUser(input: $input) { id username admin } }`,
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
        `mutation { refreshToken { user { id username admin } } }`,
      );
      user.value = data.refreshToken.user;
    } catch (err) {
      console.log(err);
    }
  }

  /** Admin-only: grant/revoke admin. Returns the updated user; re-throws. */
  async function setUserAdmin(userId, admin) {
    try {
      const data = await gql(
        `mutation SetUserAdmin($id: ID!, $admin: Boolean!) { setUserAdmin(id: $id, admin: $admin) { id username admin } }`,
        { id: userId, admin },
      );
      return data.setUserAdmin;
    } catch (err) {
      console.error(err);
      throw err;
    }
  }

  return {
    user,

    loggedIn,

    logIn,
    refreshToken,
    logOut,
    createUser,
    setUserAdmin,
  };
});
