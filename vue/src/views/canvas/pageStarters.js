/**
 * Starting layouts for a new canvas page, offered when one is created
 * (PageForm.vue). Each `boxes(title)` returns box inputs for
 * stores/boxes.js — without `page`, which the store adds. Everything is an
 * ordinary box afterwards: move it, edit it, delete it.
 */
import { galleryHtml, imageHtml } from "@/views/home3/templates";

const escape = (s) =>
  s.replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" })[c]);

const html = (title, body, geom, extra = {}) => ({ kind: "html", title, html: body, ...geom, ...extra });

function shrine(title) {
  const t = escape(title);
  return [
    html(
      "Heading",
      `<style>
  body { display: grid; place-items: center; text-align: center; }
  h1 { margin: 0; font-size: 4.5rem; line-height: 1; }
  p { margin: .3rem 0 0; color: var(--color-tertiary); }
</style>
<div>
  <h1>${t}</h1>
  <p>a shrine</p>
</div>`,
      { x: 300, y: 20, w: 800, h: 160 },
      { framed: false },
    ),
    html(
      "Main image",
      imageHtml("/img/memes/epic.jpeg"),
      { x: 300, y: 200, w: 380, h: 380 },
      { framed: false },
    ),
    html(
      "About",
      `<h3>Why ${t}</h3>
<p>What it is, and why it gets a shrine.</p>
<h3>Favourite moments</h3>
<ul>
  <li>…</li>
</ul>`,
      { x: 700, y: 200, w: 400, h: 380 },
    ),
    html(
      "Gallery",
      galleryHtml(["/img/memes/fire-woman.gif", "/img/memes/mix.gif", "/img/memes/lol.jpg"]),
      { x: 300, y: 600, w: 800, h: 320 },
    ),
    html(
      "Quote",
      `<style>
  body { display: grid; place-items: center; text-align: center; font-style: italic; }
</style>
<blockquote>“A line worth remembering.”</blockquote>`,
      { x: 300, y: 940, w: 800, h: 120 },
    ),
  ].map((box, i) => ({ z: i + 1, ...box }));
}

function notes() {
  const note = (text, x, y) =>
    html(
      "Note",
      `<style>
  body { padding: 0; }
  .note { box-sizing: border-box; height: 100%; padding: .75rem 1rem; background: #f6e27a;
    color: #2b2610; box-shadow: 2px 3px 0 rgba(0, 0, 0, .35); overflow: auto; }
  .note h3 { margin: 0 0 .3rem; color: inherit; }
</style>
<div class="note"><h3>${text}</h3><ul><li>…</li></ul></div>`,
      { x, y, w: 260, h: 220 },
      { framed: false, adminOnly: true },
    );
  return [note("To do", 40, 40), note("Ideas", 330, 40), note("Bugs", 620, 40)].map((box, i) => ({
    z: i + 1,
    ...box,
  }));
}

export const STARTERS = [
  { id: "blank", label: "Blank", boxes: () => [] },
  { id: "shrine", label: "Shrine", boxes: shrine },
  { id: "notes", label: "Notes board (admin-only notes)", boxes: notes },
];
