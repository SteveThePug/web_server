<script setup>
import { RouterView } from "vue-router";
</script>

<template>
  <div class="cv-layout">
    <RouterView />
  </div>
</template>

<!--
  Light theme for /cv and /cv/jobs.

  Unscoped on purpose. It re-points the site's colour tokens (defined in
  src/assets/styles.css) to ink-on-paper values so the global element styles
  render light, then undoes the few dark-site element styles that aren't
  colours. It also holds the button styles shared by the CV pages.
-->
<style>
.cv-layout {
  /* Paper palette */
  --color-ink: #111;
  --color-ink-soft: #333;
  --color-ink-muted: #555;
  --color-ink-faint: #999;
  --color-ink-link: #0066cc;
  --color-ink-danger: #c00;
  --color-paper: #fff;
  --color-paper-tint: #f5f5f5;
  --color-paper-shade: #eee;
  --color-line: #ccc;
  --color-line-soft: #e0e0e0;

  /* Site tokens -> paper */
  --color-primary: var(--color-ink);
  --color-secondary: var(--color-ink);
  --color-tertiary: var(--color-ink);
  --color-quaternary: var(--color-line);
  --color-muted: var(--color-ink-muted);
  --color-surface: var(--color-paper);
  --color-surface-deep: var(--color-paper);
  --color-link-bg: transparent;

  min-height: 100vh;
  background: var(--color-paper);
  color: var(--color-ink);
}

/* Undo dark-site element styling that isn't just colour */
.cv-layout h1,
.cv-layout h2,
.cv-layout h3,
.cv-layout h4 {
  margin: 0;
}
.cv-layout a {
  text-align: left;
  font-family: inherit;
  letter-spacing: normal;
}
.cv-layout input,
.cv-layout textarea {
  width: auto;
  padding: 0;
  background: var(--color-paper);
  border-color: var(--color-line);
}
.cv-layout input::placeholder,
.cv-layout textarea::placeholder {
  color: var(--color-ink-faint);
  opacity: 1;
}
.cv-layout table,
.cv-layout tr {
  border-color: transparent;
}
.cv-layout th {
  border: none;
}
.cv-layout th,
.cv-layout td {
  padding: 0;
}

/* Buttons shared by the CV selector bar and the job-applications page */
.cv-btn {
  padding: 0.4rem 1rem;
  border: 1px solid var(--color-ink-soft);
  border-radius: 4px;
  background: var(--color-paper);
  color: var(--color-ink-soft);
  cursor: pointer;
  font-size: 0.9rem;
  white-space: nowrap;
  text-decoration: none;
  transition:
    background 0.15s,
    color 0.15s;
}
.cv-btn:hover {
  background: var(--color-paper-shade);
}
.cv-btn:disabled {
  opacity: 0.5;
  cursor: default;
}
.cv-btn-sm {
  padding: 0.2rem 0.55rem;
  font-size: 0.8rem;
}
.cv-btn-primary,
.cv-btn.active {
  background: var(--color-ink-soft);
  color: var(--color-paper);
}
.cv-btn-primary:hover {
  background: var(--color-ink-muted);
}
.cv-btn-danger {
  border-color: var(--color-ink-danger);
  color: var(--color-ink-danger);
}
.cv-btn-danger:hover {
  background: #fee;
}
</style>
