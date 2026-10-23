/**
 * Geometry for dragging and resizing a box, with smart guides.
 *
 * While a box is dragged, its left edge, centre and right edge (and top,
 * middle and bottom) are compared with every other box's, and with the
 * canvas's sides, centre and top. When one comes within SNAP_PX screen px of
 * another it snaps onto it, and every line the box then shares is returned as
 * a guide for BoxCanvas to draw. A resize does the same for the edges being
 * dragged only. Away from all of them the 5px grid applies, as it always did.
 *
 * Everything here is canvas px. The threshold is in screen px because that is
 * what the hand feels; the caller divides it by the canvas scale.
 */
export const GRID = 5;
export const MIN_SIZE = 40;
export const SNAP_PX = 6;

const grid = (n) => Math.round(n / GRID) * GRID;
// Start, centre and end of a span.
const thirds = (start, size) => [start, start + size / 2, start + size];

/**
 * The lines a dragged box can snap to, per axis. Each carries the extent of
 * what it came from on the other axis, so a guide can be drawn between the
 * two boxes rather than across the whole canvas. Computed once per drag.
 */
export function snapTargets(others, canvasW) {
  const x = thirds(0, canvasW).map((at) => ({ at, from: 0, to: 0 }));
  // The canvas is as tall as its lowest box, so only its top is a fixed line.
  const y = [{ at: 0, from: 0, to: canvasW }];
  for (const b of others) {
    for (const at of thirds(b.x, b.w)) x.push({ at, from: b.y, to: b.y + b.h });
    for (const at of thirds(b.y, b.h)) y.push({ at, from: b.x, to: b.x + b.w });
  }
  return { x, y };
}

export const NO_TARGETS = { x: [], y: [] };

// Which of the box's edges a gesture moves: all four for a move, otherwise
// the ones named by the handle ("ne", "w", …). Spelled out because "move"
// itself contains an "e".
function edgesOf(mode) {
  if (mode === "move") return { w: true, e: true, n: true, s: true };
  return { w: mode.includes("w"), e: mode.includes("e"), n: mode.includes("n"), s: mode.includes("s") };
}

// The lines of `g` that the gesture is moving, per axis. A move carries the
// centres along too; a resize only moves the edge under the pointer.
function movingLines(g, mode) {
  if (mode === "move") return { x: thirds(g.x, g.w), y: thirds(g.y, g.h) };
  const edges = edgesOf(mode);
  const x = [];
  const y = [];
  if (edges.w) x.push(g.x);
  if (edges.e) x.push(g.x + g.w);
  if (edges.n) y.push(g.y);
  if (edges.s) y.push(g.y + g.h);
  return { x, y };
}

// The smallest shift that puts one of the moving lines on a target, or null
// if none is within `threshold`.
function nearest(moving, targets, threshold) {
  let best = null;
  for (const m of moving) {
    for (const t of targets) {
      const d = t.at - m;
      if (Math.abs(d) <= threshold && (best === null || Math.abs(d) < Math.abs(best))) best = d;
    }
  }
  return best;
}

// One axis of the result. `lo` and `hi` say whether the gesture moves the
// box's start and end edge on this axis; `shift` is the snap from nearest(),
// or null to fall back to the grid.
function resolve(start, size, delta, lo, hi, shift) {
  const fit = (n) => (shift === null ? grid(n) : Math.round(n + shift));
  if (lo && hi) return [Math.max(0, fit(start + delta)), size];
  if (hi) return [start, Math.max(MIN_SIZE, fit(size + delta))];
  if (lo) {
    const s = Math.max(0, Math.min(fit(start + delta), start + size - MIN_SIZE));
    return [s, start + size - s];
  }
  return [start, size];
}

/**
 * The geometry for a drag of (dx, dy) canvas px from `from`, in `mode`
 * ("move" or a handle), snapped to `targets` within `threshold` canvas px.
 * Pass NO_TARGETS for the grid alone.
 */
export function place(from, mode, dx, dy, targets, threshold) {
  const edges = edgesOf(mode);
  // Where the pointer alone would put the box, before any snapping.
  const raw = { ...from };
  if (edges.w) raw.x += dx;
  if (edges.e !== edges.w) raw.w += edges.e ? dx : -dx;
  if (edges.n) raw.y += dy;
  if (edges.s !== edges.n) raw.h += edges.s ? dy : -dy;

  const moving = movingLines(raw, mode);
  const sx = nearest(moving.x, targets.x, threshold);
  const sy = nearest(moving.y, targets.y, threshold);
  const [x, w] = resolve(from.x, from.w, dx, edges.w, edges.e, sx);
  const [y, h] = resolve(from.y, from.h, dy, edges.n, edges.s, sy);
  return { x, y, w, h };
}

// Each target line that one of the moving lines now sits on, spanning both
// the box (from..to) and whatever the line came from. Lines from several
// boxes at the same place merge into one guide.
function matches(moving, targets, from, to) {
  const found = new Map();
  for (const t of targets) {
    // Within a px, since centres of odd sizes fall on half px and the
    // geometry is rounded.
    if (!moving.some((m) => Math.abs(t.at - m) < 1)) continue;
    const [lo, hi] = found.get(t.at) ?? [from, to];
    found.set(t.at, [Math.min(lo, t.from), Math.max(hi, t.to)]);
  }
  return [...found].map(([at, [lo, hi]]) => ({ at, from: lo, to: hi }));
}

/** The guides to draw for box geometry `g` placed by place(). */
export function guidesFor(g, mode, targets) {
  const moving = movingLines(g, mode);
  return [
    ...matches(moving.x, targets.x, g.y, g.y + g.h).map((m) => ({ axis: "x", ...m })),
    ...matches(moving.y, targets.y, g.x, g.x + g.w).map((m) => ({ axis: "y", ...m })),
  ];
}
