/**
 * The stylesheet injected at the top of every HTML box's iframe, so admin
 * HTML looks like the rest of the site without restating the palette.
 *
 * Two things a sandboxed iframe cannot do for itself:
 *   - Inherit CSS. It is a separate document, so the design tokens are copied
 *     in as literal values, read from the live page so a re-themed layout
 *     (CVLayout re-points the tokens) would carry through.
 *   - Load our fonts. Without allow-same-origin its origin is opaque ("null"),
 *     which makes /fonts/*.woff2 a cross-origin font load that nginx and the
 *     Vite dev server would both refuse without CORS headers. Instead the page
 *     fetches the two files (28 KB) once, same-origin, and inlines them as
 *     data: URLs.
 */

const TOKENS = [
  "--color-primary",
  "--color-secondary",
  "--color-tertiary",
  "--color-quaternary",
  "--color-muted",
  "--color-surface",
  "--color-surface-deep",
  "--color-surface-tint",
  "--color-link-bg",
];

const FONTS = [
  ["big_noodle_titling", "/fonts/big_noodle_titling.woff2"],
  ["CreatoDisplay", "/fonts/CreatoDisplay-Bold.woff2"],
];

async function toDataUrl(url) {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`${url}: ${res.status}`);
  const blob = await res.blob();
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result);
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(blob);
  });
}

async function fontFaces() {
  // A font that fails to load falls back to sans-serif rather than holding
  // back every box on the page.
  const faces = await Promise.all(
    FONTS.map(async ([family, url]) => {
      try {
        const src = await toDataUrl(url);
        return `@font-face{font-family:"${family}";src:url(${src}) format("woff2");font-display:swap}`;
      } catch {
        return "";
      }
    }),
  );
  return faces.join("\n");
}

function tokens() {
  const style = getComputedStyle(document.documentElement);
  const decls = TOKENS.map((t) => `${t}:${style.getPropertyValue(t).trim()};`);
  return `:root{${decls.join("")}--font-heading:"big_noodle_titling",sans-serif;--font-body:"CreatoDisplay",sans-serif}`;
}

const BASE = `
html,body{margin:0;height:100%}
body{box-sizing:border-box;padding:8px;overflow:auto;background:transparent;
  color:var(--color-secondary);font-family:var(--font-body);font-size:15px}
*,*::before,*::after{box-sizing:inherit}
h1,h2,h3,h4{font-family:var(--font-heading);color:var(--color-primary);margin:0 0 .25em;font-weight:normal;letter-spacing:.025em}
a{color:var(--color-primary)}
img{max-width:100%}
`;

let cached = null;

/** Resolves to the <style> body for an HTML box. Built once per page load. */
export function frameCss() {
  cached ??= fontFaces().then((faces) => `${faces}\n${tokens()}\n${BASE}`);
  return cached;
}

/** A complete srcdoc document wrapping `html` in the site styles. */
export function buildSrcdoc(css, html) {
  return `<!doctype html><html><head><meta charset="utf-8"><style>${css}</style></head><body>${html}</body></html>`;
}
