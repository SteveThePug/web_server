import { createRouter, createWebHistory } from "vue-router";
import Home from "../views/Home.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: Home,
    },
    {
      path: "/cv",
      name: "cv",
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import("../views/CV.vue"),
    },
    {
      path: "/admin",
      name: "admin",
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import("../views/Admin.vue"),
    },
    {
      path: "/bookmarks",
      name: "bookmarks",
      component: () => import("../views/Bookmarks.vue"),
    },
    {
      path: "/notes",
      name: "notes",
      component: () => import("../views/Notes.vue"),
    },
    {
      path: "/notes/islp",
      name: "islp",
      component: () => import("../views/notes/Islp.vue"),
    },
    {
      path: "/notes/vue",
      name: "vue",
      component: () => import("../views/notes/Vue.vue"),
    },
    {
      path: "/shrines",
      name: "shrine links",
      component: () => import("../views/Shrines.vue"),
    },
    {
      path: "/shrines/gto",
      name: "gto shrine",
      component: () => import("../views/shrines/GTO.vue"),
    },
    {
      path: "/shrines/skipskipbenben",
      name: "skipskipbenben shrine",
      component: () => import("../views/shrines/Skipskipbenben.vue"),
    },
    {
      path: "/shrines/evangelion",
      name: "evangelion shrine",
      component: () => import("../views/shrines/Evangelion.vue"),
    },
    {
      path: "/shrines/demoman",
      name: "demoman shrine",
      component: () => import("../views/shrines/Demoman.vue"),
    },
    {
      path: "/:pathMatch(.*)*",
      name: "404",
      component: () => import("../views/404.vue"),
    },
  ],
});

export default router;
