import { defineStore } from "pinia";
import { computed, ref, watch } from "vue";
import { gql } from "@/graphql";
import { useHomeDataStore } from "@/stores/homeData";

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

  const homeData = useHomeDataStore();
  watch(
    () => homeData.posts,
    (newPosts) => {
      if (newPosts.length > 0) {
        posts.value = newPosts;
      }
    },
    { immediate: true },
  );

  async function fetchPosts() {
    await homeData.fetchAll();
  }

  async function deletePost(post) {
    try {
      await gql(
        `mutation DeletePost($id: ID!) { deletePost(id: $id) { id } }`,
        { id: post.id },
      );
      console.log("Deleted:", post.id);
      await homeData.fetchAll();
    } catch (err) {
      console.error("Delete failed:", err);
    }
  }

  return {
    posts,

    postsCount,

    fetchPosts,
    deletePost,
  };
});
