/**
 * The pre-Home3 /stp page (views/home/Home.vue) redrawn as boxes, for the
 * one-off "Import current layout" button on an empty canvas.
 *
 * Home.vue was an A4 sheet (~800×1120px) holding a 10×10 CSS grid with 5px
 * gaps, between two 240px sidebars. The grid areas below are copied from its
 * grid-template-areas and turned into px here so they can be read against the
 * original; the sidebars are stacked top to bottom.
 */
import { imageHtml } from "./templates";

const SHEET = { x: 300, y: 20, w: 800, h: 1120 };
const GAP = 5;
const CELL_W = (SHEET.w - 9 * GAP) / 10;
const CELL_H = (SHEET.h - 9 * GAP) / 10;

// [widget, column, row, column span, row span] — zero-based, as in Home.vue.
const GRID = [
  ["Intro", 0, 0, 6, 4],
  ["Listening", 6, 0, 4, 3],
  ["Stamps", 6, 3, 4, 1],
  ["Feed", 0, 4, 3, 4],
  ["Links", 3, 4, 2, 4],
  ["Collage", 5, 4, 5, 4],
  ["Consumption", 0, 8, 4, 2],
  ["Gym", 4, 8, 3, 2],
  ["Favorites", 7, 8, 3, 2],
];

// Collage and Gym were hidden on phones; keep that.
const HIDDEN_ON_MOBILE = new Set(["Collage", "Gym"]);

function gridBox([widget, col, row, cols, rows]) {
  return {
    kind: "widget",
    widget,
    x: Math.round(SHEET.x + col * (CELL_W + GAP)),
    y: Math.round(SHEET.y + row * (CELL_H + GAP)),
    w: Math.round(cols * CELL_W + (cols - 1) * GAP),
    h: Math.round(rows * CELL_H + (rows - 1) * GAP),
    hideOnMobile: HIDDEN_ON_MOBILE.has(widget),
  };
}

// [widget or html box, height]; each column fills SHEET.h top to bottom.
function column(x, entries) {
  let y = SHEET.y;
  return entries.map(([box, h]) => {
    const placed = { ...box, x, y, w: 240, h };
    y += h + GAP;
    return placed;
  });
}

const widget = (name, extra = {}) => ({ kind: "widget", widget: name, ...extra });

export function defaultLayout() {
  const left = column(50, [
    [widget("Time", { hideOnMobile: true }), 90],
    [widget("Timer", { hideOnMobile: true }), 130],
    [widget("Radio", { hideOnMobile: true }), 160],
    [widget("Commits"), 487],
    [
      {
        kind: "html",
        title: "Fire woman",
        html: imageHtml("/img/memes/fire-woman.gif"),
        framed: false,
        hideOnMobile: true,
      },
      238,
    ],
  ]);
  const right = column(1110, [
    [widget("Steam"), 205],
    [widget("Bookmarks"), 250],
    [widget("Chat"), 410],
    [widget("Miku", { framed: false, hideOnMobile: true }), 240],
  ]);
  return [...GRID.map(gridBox), ...left, ...right].map((box, i) => ({ z: i + 1, ...box }));
}
