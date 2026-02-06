import { defineStore } from "pinia";
import { computed, ref } from "vue";
import axios from "axios";

const post_template = {
  title: "Can't fetch from the db yo",
  content:
    "This is meant to be pulling from a database, but for some reason that isn't working and this is filler text that should hopefully never see the light of day. If you are reading this, something has gone horribly, horribly wrong. Please start crying and prepare for the incoming wrath of hell. Furthermore, this is very, very long because I am trying to test the scroll feature so thank you ^_^.",
  author: {
    username: "stp",
  },
  createdAt: Date.now(),
};

export const usePostsStore = defineStore("posts", () => {
  const posts = ref([post_template]);

  const postsCount = computed(() => posts.value.length);

  async function fetchPosts() {
    try {
      const res = await axios.get("/api/posts");
      if (!Array.isArray(res.data)) {
        throw new Error("Invalid response from posts API");
      }
      posts.value = res.data;
    } catch (err) {
      console.error("Cannot connect to Post API", err);
    }
  }

  async function deletePost(post) {
    try {
      const res = await axios.delete(
        `/api/posts/${encodeURIComponent(post.id)}`,
      );
      console.log("Deleted:", res.data);
      fetchPosts();
    } catch (err) {
      console.error("Delete failed:", err);
    }
  }

  fetchPosts();

  return {
    posts,

    postsCount,

    fetchPosts,
    deletePost,
  };
});
