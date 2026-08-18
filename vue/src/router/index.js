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
