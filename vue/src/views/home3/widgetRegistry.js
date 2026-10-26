/**
 * The built-in Vue widgets a "widget" box can show, by the name stored in
 * Box.widget. Renaming a key orphans every box saved under the old name
 * (WidgetBox shows a placeholder for it), so treat the keys as stored data.
 *
 * Each is an async component so a widget nobody has placed on the canvas
 * never loads. `size` is what a newly added box of that kind starts at.
 */
import { defineAsyncComponent } from "vue";

const lazy = (load) => defineAsyncComponent(load);

export const WIDGETS = {
  Intro: { component: lazy(() => import("@/views/home/Intro3.vue")), size: [480, 440] },
  Listening: { component: lazy(() => import("@/views/home/Listening.vue")), size: [310, 330] },
  Stamps: { component: lazy(() => import("@/views/home/Stamps.vue")), size: [310, 110] },
  Feed: { component: lazy(() => import("@/views/home/Feed.vue")), size: [235, 440] },
  Links: { component: lazy(() => import("@/views/home/Links.vue")), size: [155, 440] },
  Collage: { component: lazy(() => import("@/views/home/Collage.vue")), size: [395, 440] },
  Places: { component: lazy(() => import("@/views/home/Places.vue")), size: [395, 440] },
  Consumption: { component: lazy(() => import("@/views/home/Consumption.vue")), size: [315, 220] },
  Favorites: { component: lazy(() => import("@/views/home/Favorites.vue")), size: [235, 220] },
  Gym: { component: lazy(() => import("@/views/home/Gym2.vue")), size: [235, 220] },
  Steam: { component: lazy(() => import("@/views/home/Steam.vue")), size: [240, 205] },
  Bookmarks: { component: lazy(() => import("@/views/home/Bookmarks.vue")), size: [240, 250] },
  Miku: { component: lazy(() => import("@/views/home/Miku.vue")), size: [240, 240] },
  Time: { component: lazy(() => import("@/components/util/Time.vue")), size: [240, 90] },
  Timer: { component: lazy(() => import("@/components/util/Timer.vue")), size: [240, 130] },
  Radio: { component: lazy(() => import("@/components/util/Radio.vue")), size: [240, 160] },
  Chat: { component: lazy(() => import("@/components/util/Chat.vue")), size: [240, 450] },
  PrivateChat: { component: lazy(() => import("@/components/util/PrivateChat.vue")), size: [240, 450] },
  Commits: { component: lazy(() => import("@/components/util/CommitHistory.vue")), size: [240, 400] },
};
