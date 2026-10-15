/**
 * Starting layouts for a new canvas page, offered when one is created
 * (PageForm.vue). Each `boxes(title)` returns box inputs for
 * stores/boxes.js — without `page`, which the store adds. Everything is an
 * ordinary box afterwards: move it, edit it, delete it.
 */
const html = (title, body, geom, extra = {}) => ({ kind: "html", title, html: body, ...geom, ...extra });

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
  { id: "notes", label: "Notes board (admin-only notes)", boxes: notes },
];
