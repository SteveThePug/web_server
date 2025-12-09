import { defineStore } from "pinia";
import { computed, ref } from "vue";
import axios from "axios";

export const useAuthStore = defineStore("auth", () => {
  const user = ref({});

  function logOut() {
    user.value = {};
  }

  async function logIn(username, password) {
    try {
      const res = await axios.post("/api/auth/login", {
        username,
        password,
      });
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
      console.log(err);
    }
  }
});
