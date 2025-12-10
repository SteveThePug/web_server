import { defineStore } from "pinia";
import { computed, ref } from "vue";
import axios from "axios";

export const useAuthStore = defineStore("auth", () => {
  const user = ref({});
  checkToken();

  const loggedIn = computed(() => !!user.username);

  async function logOut() {
    try {
      const res = await axios.post("/api/auth/logout");
    } catch (err) {
      console.error(err);
    }
    user.value = {};
  }

  async function logIn(username, password) {
    try {
      const res = await axios.post("/api/auth/login", {
        username,
        password,
      });
      user.value = res.data;
    } catch (err) {
      console.error(err);
    }
  }

  async function createUser(username, password) {
    try {
      const res = await axios.post("/api/user", {
        username,
        password,
      });
      user.value = res.data;
    } catch (err) {
      console.error(err);
    }
  }

  async function refreshToken() {
    try {
      const res = await axios.post("/api/auth/refresh");
      user.value = res.data;
    } catch (err) {
      console.log(err);
    }
  }

  async function checkToken() {
    try {
      const res = await axios.get("/api/auth/check");
      user.value = res.data;
    } catch (err) {
      user.value = {};
    }
  }

  return {
    user,
    checkToken,
    logIn,
    refreshToken,
    logOut,
    createUser,
    loggedIn,
  };
});
