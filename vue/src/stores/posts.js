/**
 * Blog posts for the Feed widget. A view over homeData.posts, built by
 * createMirrorStore(); see stores/createMirrorStore.js for the
 * placeholder-then-overwrite pattern.
 *
 * `post_template` is placeholder content rendered before (or instead of) a
 * successful fetch, so the widget never flashes empty.
 */

import { defineStore } from "pinia";
import { gql } from "@/graphql";
import { createMirrorStore } from "@/stores/createMirrorStore";

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
  const { items, count, fetch, homeData } = createMirrorStore({
    slice: "posts",
    template: post_template,
  });

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
    posts: items,

    postsCount: count,

    fetchPosts: fetch,
    deletePost,
  };
});
