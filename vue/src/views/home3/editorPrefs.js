/**
 * The box editor panel's size and its code's text size, shared between
 * BoxEditor (which draws the panel and its controls) and BoxCanvas (which
 * pads itself by the panel's width so the canvas rescales beside it rather
 * than under it), and whether dragged boxes snap to each other (BoxCanvas's
 * Align button). They are one admin's preference for their own screen, so
 * they live in localStorage, not the database.
 */
import { ref, watch } from "vue";

const KEY = "boxEditor";

export const PANEL_WIDTH = { min: 320, default: 460 };
export const FONT_SIZE = { min: 9, max: 24, default: 13 };

function load() {
  try {
    return JSON.parse(localStorage.getItem(KEY)) ?? {};
  } catch {
    return {}; // storage unavailable or corrupted: keep the defaults
  }
}

const saved = load();
const clamp = (n, min, max) => Math.min(max, Math.max(min, n));
const valid = (n) => typeof n === "number" && Number.isFinite(n);

// Not clamped to the window here: the panel's CSS caps it with min(…, 100vw),
// so a width chosen on a wide monitor survives a visit from a narrow one.
export const panelWidth = ref(valid(saved.width) ? Math.max(PANEL_WIDTH.min, saved.width) : PANEL_WIDTH.default);
export const fontSize = ref(valid(saved.fontSize) ? clamp(saved.fontSize, FONT_SIZE.min, FONT_SIZE.max) : FONT_SIZE.default);
export const snapping = ref(saved.snapping !== false);

export function setPanelWidth(width) {
  panelWidth.value = Math.round(clamp(width, PANEL_WIDTH.min, window.innerWidth));
}

/** Step the code's text size by `delta` px; 0 resets it. */
export function zoom(delta) {
  fontSize.value = delta === 0 ? FONT_SIZE.default : clamp(fontSize.value + delta, FONT_SIZE.min, FONT_SIZE.max);
}

watch([panelWidth, fontSize, snapping], ([width, size, snap]) => {
  try {
    localStorage.setItem(KEY, JSON.stringify({ width, fontSize: size, snapping: snap }));
  } catch {
    /* storage unavailable: the preference just won't outlive the tab */
  }
});
