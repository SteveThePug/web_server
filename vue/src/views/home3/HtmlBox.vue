<script setup>
/**
 * An admin-authored HTML box, rendered in a sandboxed iframe.
 *
 * The sandbox is the security boundary for the whole "write HTML in the
 * browser" feature. `allow-scripts` without `allow-same-origin` gives the frame
 * an opaque origin: its scripts run, but cannot read our cookies, call /api as
 * the visitor, or reach into the parent page. NEVER add allow-same-origin —
 * together with allow-scripts it lets the frame remove its own sandbox.
 *
 * The popup and top-navigation flags only let links work: target="_blank"
 * opens a normal (unsandboxed) tab, and target="_top" can navigate the page on
 * a real click.
 *
 * Nothing is rendered until the frame stylesheet (fonts inlined, see
 * frameStyles.js) is ready, so boxes don't flash in fallback fonts.
 */
import { computed, ref } from "vue";
import { buildSrcdoc, frameCss } from "./frameStyles";

const props = defineProps({
  html: { type: String, default: "" },
  title: { type: String, default: "" },
});

const css = ref(null);
frameCss().then((value) => (css.value = value));

const srcdoc = computed(() =>
  css.value === null ? null : buildSrcdoc(css.value, props.html),
);
</script>

<template>
  <iframe
    v-if="srcdoc !== null"
    class="html-box"
    :srcdoc="srcdoc"
    :title="title || 'Custom box'"
    sandbox="allow-scripts allow-popups allow-popups-to-escape-sandbox allow-top-navigation-by-user-activation"
    referrerpolicy="no-referrer"
    loading="lazy"
  />
</template>

<style scoped>
.html-box {
  display: block;
  width: 100%;
  height: 100%;
  border: 0;
  background: transparent;
}
</style>
