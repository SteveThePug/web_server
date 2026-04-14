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
