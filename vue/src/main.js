/**
 * App entry point. Mounts <App> on #app with the router and a fresh Pinia.
 *
 * The top-level `await init()` loads the Rust/wasm bundle before anything else
 * renders, so components that reach for a wasm export (components/util/wasm/)
 * never see an uninitialised module. Top-level await in an entry module is what
 * `vite-plugin-top-level-await` in vite.config.js is there to support.
 */

import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import router from "./router";
import "./assets/styles.css";
import init from "@/wasm/stp_wasm.js";

await init();

const app = createApp(App);

app.use(router);
app.use(createPinia());

app.mount("#app");
