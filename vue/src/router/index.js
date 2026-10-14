/**
 * Route table and the single admin route guard.
 *
 * Structure: two layout shells. `/` mounts DefaultLayout (navbar + footer, dark
 * theme) and holds most routes; `/cv` mounts CVLayout (light print theme).
 * `/stp2` is deliberately outside both — it is a bare full-bleed mock.
 *
 * Pages made in the browser (views/canvas/) have no entries here: the
 * catch-all at the end looks the path up and renders the page if one exists.
 *
 * Only the landing page is statically imported; every other view is a dynamic
 * import so Vite code-splits it into its own chunk.
 *
 * Route `name`s are public contracts (used by RouterLink/`router.push`) — don't
 * rename them.
 */

import { createRouter, createWebHistory } from "vue-router";
import DefaultLayout from "@/layouts/DefaultLayout.vue";
import CVLayout from "@/layouts/CVLayout.vue";
import Landing from "@/views/landing/Landing.vue";
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
                    component: () => import("@/views/home3/Home3.vue"),
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
                // The GTO, Evangelion and Skipskipbenben shrines were <Wip />
                // placeholders; they are canvas pages now, made from /shrines
                // and served by the catch-all below.
                {
                    path: "shrines/demoman",
                    name: "demoman shrine",
                    component: () => import("@/views/home/shrines/Demoman.vue"),
                },
                // Anything else is an admin-created canvas page if one exists
                // at that address (the Page GraphQL type), else the 404.
                {
                    path: ":pathMatch(.*)*",
                    name: "404",
                    component: () => import("@/views/canvas/CanvasPage.vue"),
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
                // /cv/jobs and /cv/leads moved to the job tracker app at
                // /jobs/; nginx redirects the old paths.
            ],
        },
    ],
});

// Guard for routes marked `meta: { requiresAdmin: true }`.
//
// The admin flag arrives asynchronously, so navigating straight to /admin on a
// cold load would otherwise read an empty user and bounce a legitimate admin to
// the login page. The guard therefore awaits auth.ensureLoaded(), which issues
// a standalone `me` query unless the user is already known.
//
// It deliberately does not wait on the home query (stores/homeData.js): that
// one fans out to Spotify, Steam and Gitea, and a guard that waited on it hung
// forever when it failed. ensureLoaded() always settles — a failed `me` reads
// as "not signed in" — and `user` is re-read on every navigation, so a login or
// logout during the session is honoured.
router.beforeEach(async (to) => {
    if (!to.meta.requiresAdmin) return;
    const auth = useAuthStore();
    await auth.ensureLoaded();
    if (!auth.user.admin)
        return { path: "/admin/login", query: { redirect: to.fullPath } };
});

export default router;
