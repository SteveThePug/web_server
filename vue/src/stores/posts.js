/**
 * Blog posts for the Feed widget. A view over homeData.posts.
 *
 * `post_template` is placeholder content rendered before (or instead of) a
 * successful fetch — the watch only overwrites `posts` when the fetch returned a
 * non-empty list, so the widget never flashes empty.
 */

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
  // Mirror the shared home query. `immediate` matters: homeData may already
  // have loaded by the time this store is first used, and a plain watch would
  // never fire for that existing value. The `length > 0` test keeps the
  // placeholder on screen when the backend returns nothing.
  watch(
    () => homeData.posts,
    (newPosts) => {
      if (newPosts.length > 0) {
        posts.value = newPosts;
      }
    },
    { immediate: true },
  );

  /** Refresh by re-running the whole home query; the watch above applies it. */
  async function fetchPosts() {
    await homeData.fetchAll();
  }

  /** Soft-delete a post, then refetch so every widget sees the new list. */
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
