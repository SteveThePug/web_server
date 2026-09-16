/**
 * Route table and the single admin route guard.
 *
 * Structure: two layout shells. `/` mounts DefaultLayout (navbar + footer, dark
 * theme) and holds most routes; `/cv` mounts CVLayout (light print theme).
 * `/stp2` is deliberately outside both — it is a bare full-bleed mock.
 *
 * Only the landing page is statically imported; every other view is a dynamic
 * import so Vite code-splits it into its own chunk.
 *
 * Route `name`s are public contracts (used by RouterLink/`router.push`) — don't
 * rename them.
 */

import { createRouter, createWebHistory } from "vue-router";
import { watch } from "vue";
import DefaultLayout from "@/layouts/DefaultLayout.vue";
import CVLayout from "@/layouts/CVLayout.vue";
import Landing from "@/views/landing/Landing.vue";
import { useHomeDataStore } from "@/stores/homeData";
import { useAuthStore } from "@/stores/auth";

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: "/",
            component: DefaultLayout,
            children: [
                {
                    path: "",
                    name: "landing",
                    component: Landing,
                },
                {
                    path: "stp",
                    name: "home",
                    component: () => import("@/views/home/Home.vue"),
                },
                {
                    path: "admin/login",
                    name: "admin-login",
                    component: () => import("@/views/admin/Login.vue"),
                },
                {
                    path: "admin",
                    name: "admin",
                    component: () => import("@/views/admin/Admin.vue"),
                    meta: { requiresAdmin: true },
                },
                {
                    path: "automata",
                    name: "mobile-automata",
                    component: () =>
                        import("@/views/automata/MobileAutomata.vue"),
                },
                {
                    path: "hotels",
                    name: "cheap-hotels",
                    component: () => import("@/views/hotels/CheapHotels.vue"),
                },
                {
                    path: "shrines",
                    name: "shrine links",
                    component: () => import("@/views/home/shrines/Shrines.vue"),
                },
                {
                    path: "shrines/gto",
                    name: "gto shrine",
                    component: () => import("@/views/home/shrines/GTO.vue"),
                },
                {
                    path: "shrines/skipskipbenben",
                    name: "skipskipbenben shrine",
                    component: () =>
                        import("@/views/home/shrines/Skipskipbenben.vue"),
                },
                {
                    path: "shrines/evangelion",
                    name: "evangelion shrine",
                    component: () =>
                        import("@/views/home/shrines/Evangelion.vue"),
                },
                {
                    path: "shrines/demoman",
                    name: "demoman shrine",
                    component: () => import("@/views/home/shrines/Demoman.vue"),
                },
                {
                    path: ":pathMatch(.*)*",
                    name: "404",
                    component: () => import("@/views/404/404.vue"),
                },
            ],
        },
        {
            path: "/stp2",
            name: "home2",
            component: () => import("@/views/home2/Home2.vue"),
        },
        {
            path: "/cv",
            component: CVLayout,
            children: [
                {
                    path: "",
                    name: "cv",
                    component: () => import("@/views/CV/CV.vue"),
                },
                {
                    path: "jobs",
                    name: "job-applications",
                    component: () => import("@/views/CV/JobApplications.vue"),
                    meta: { requiresAdmin: true },
                },
            ],
        },
    ],
});

// Guard for routes marked `meta: { requiresAdmin: true }`.
//
// The admin flag comes from homeData's `me` field, which arrives asynchronously.
// Navigating straight to /admin on a cold load would otherwise read an empty
// user and bounce a legitimate admin to the login page, so the guard first
// blocks on homeData.loaded. The watcher is self-stopping: `stop()` is called
// from inside the callback before resolving, so the promise settles exactly
// once and the watcher does not leak across navigations.
router.beforeEach(async (to) => {
    if (!to.meta.requiresAdmin) return;
    const homeData = useHomeDataStore();
    if (!homeData.loaded) {
        await new Promise((resolve) => {
            const stop = watch(
                () => homeData.loaded,
                (val) => {
                    if (val) {
                        stop();
                        resolve();
                    }
                },
            );
        });
    }
    if (!useAuthStore().user.admin)
        return { path: "/admin/login", query: { redirect: to.fullPath } };
});

export default router;
