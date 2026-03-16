import { defineStore } from "pinia";
import { computed, ref, watch } from "vue";
import { gql } from "@/graphql";
import { useHomeDataStore } from "@/stores/homeData";

export const useAuthStore = defineStore("auth", () => {
  const user = ref({});
  const loggedIn = computed(() => !!user.value.username);

  const homeData = useHomeDataStore();
  watch(
    () => homeData.me,
    (me) => {
      if (me) {
        user.value = me;
      }
    },
    { immediate: true },
  );

  async function logOut() {
    try {
      await gql(`mutation { logout }`);
    } catch (err) {
      console.error(err);
    }
    user.value = {};
  }

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
