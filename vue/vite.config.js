import { fileURLToPath, URL } from "node:url";

import { defineConfig, loadEnv } from "vite";
import tailwindcss from "@tailwindcss/vite";
import vue from "@vitejs/plugin-vue";
import vueDevTools from "vite-plugin-vue-devtools";

// Dev-server proxy targets (frontend-only `npm run dev`).
//
// In production, and in the full compose stack, nginx does this routing. The
// Vite dev server has to stand in for it, which means reaching each service
// FROM THE HOST — and the compose stack publishes almost none of them:
//
//   published by docker-compose.yml:  nginx (80/443), gitea (3000 + 2222 ssh),
//                                     the liquidsoap harbor port
//   NOT published:                    backend (8080), python (8000),
//                                     icecast (8000), db, silverbullet
//
// So the defaults below are not "where the stack is". Only the Gitea one
// matches a published port. The rest are each service's own default port, i.e.
// where it listens if you run it directly on the host (`go run .` in backend/,
// `uvicorn app.main:app` in python/) or publish its port yourself.
//
// /py and /radio share a default of :8000 because BOTH services default to
// 8000 inside their own containers (PYTHON_PORT, ICECAST_PORT). On one host
// only one process can hold that port, so at most one of the two defaults is
// ever right — in practice /py, since a host uvicorn is the usual way to work
// on the Python API and Icecast is rarely run outside compose. With nothing
// (or uvicorn) behind /radio the stream probe just fails, and the UI already
// reads that as "radio offline".
//
// Override any of them in the shell or in vue/.env.local:
//
//   VITE_PROXY_API    Go backend            default http://localhost:8080
//   VITE_PROXY_GITEA  Gitea                 default http://localhost:3000
//   VITE_PROXY_PY     FastAPI (uvicorn)     default http://localhost:8000
//   VITE_PROXY_RADIO  Icecast               default http://localhost:8000
//
// To borrow a running compose stack instead of running services on the host,
// point API / GITEA / RADIO at its nginx (e.g. VITE_PROXY_API=http://localhost):
// those three are forwarded with their path intact, which is what nginx
// expects. VITE_PROXY_PY cannot be pointed at nginx — see the rewrite below.
const DEFAULT_PROXY = {
  VITE_PROXY_API: "http://localhost:8080",
  VITE_PROXY_GITEA: "http://localhost:3000",
  VITE_PROXY_PY: "http://localhost:8000",
  VITE_PROXY_RADIO: "http://localhost:8000",
};

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // loadEnv reads the .env files (config runs before Vite loads them itself);
  // process.env is merged on top so a one-off `VITE_PROXY_PY=… npm run dev`
  // wins over a file.
  const env = {
    ...DEFAULT_PROXY,
    ...loadEnv(mode, process.cwd(), "VITE_PROXY_"),
  };
  for (const key of Object.keys(DEFAULT_PROXY))
    if (process.env[key]) env[key] = process.env[key];

  return {
    plugins: [
      vue(),
      ...(process.env.NODE_ENV !== "production" ? [vueDevTools()] : []),
      tailwindcss(),
    ],
    resolve: {
      alias: {
        "@": fileURLToPath(new URL("./src", import.meta.url)),
      },
    },
    server: {
      host: "0.0.0.0",
      proxy: {
        "/api": env.VITE_PROXY_API,
        "/gitea": env.VITE_PROXY_GITEA,
        "/radio": env.VITE_PROXY_RADIO,
        // Python API: nginx strips the /py prefix in production, so do the
        // same here. That makes the target a bare uvicorn — pointed at nginx
        // the request would arrive without the prefix nginx routes on.
        "/py": {
          target: env.VITE_PROXY_PY,
          rewrite: (path) => path.replace(/^\/py/, ""),
        },
      },
    },
  };
});
