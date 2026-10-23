<script setup>
/**
 * CodeMirror 6 HTML editor with v-model, used by BoxEditor.
 *
 * CodeMirror is imported dynamically on mount, so the ~150 KB of editor code
 * is only downloaded by an admin who opens a box for editing — never by a
 * visitor to /stp.
 *
 * The document is owned by CodeMirror; `modelValue` is pushed in only when it
 * differs from what the editor already holds, otherwise every keystroke would
 * round-trip through the parent and reset the cursor.
 *
 * `fontSize` scales the code. Ctrl/Cmd with +, - or 0 inside the editor asks
 * the parent to change it (`zoom` with +1, -1 or 0 for reset) instead of
 * zooming the whole page.
 */
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";

const props = defineProps({
  modelValue: { type: String, default: "" },
  fontSize: { type: Number, default: 13 },
});
const emit = defineEmits(["update:modelValue", "save", "zoom"]);

const host = ref(null);
const failed = ref(false);
let view = null;

onMounted(async () => {
  let mods;
  try {
    mods = await Promise.all([import("codemirror"), import("@codemirror/lang-html"), import("@codemirror/view")]);
  } catch {
    failed.value = true;
    return;
  }
  const [{ basicSetup, EditorView }, { html }, { keymap }] = mods;
  if (!host.value) return; // unmounted while loading

  const theme = EditorView.theme(
    {
      "&": { height: "100%", fontSize: "var(--code-font-size)", backgroundColor: "var(--color-surface-deep)", color: "var(--color-muted)" },
      ".cm-content": { caretColor: "var(--color-primary)" },
      ".cm-gutters": { backgroundColor: "var(--color-surface)", color: "var(--color-quaternary)", border: "none" },
      // The selection is drawn in a layer *behind* the text, so an opaque
      // active-line colour would hide it on the cursor's line: keep it
      // translucent.
      ".cm-activeLine": { backgroundColor: "#55ffbb0d" },
      ".cm-activeLineGutter": { backgroundColor: "var(--color-surface-tint)" },
      // Same selectors as CodeMirror's base theme — a shorter one loses to
      // its focused rule, which paints #233, all but invisible here.
      "& .cm-selectionBackground, &.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground": {
        backgroundColor: "#1f6b5c",
      },
      ".cm-selectionMatch": { backgroundColor: "#55ffbb26" },
      ".cm-scroller": { fontFamily: "ui-monospace, monospace" },
    },
    { dark: true },
  );

  view = new EditorView({
    parent: host.value,
    doc: props.modelValue,
    extensions: [
      // Before basicSetup so Mod-s wins over the browser's "save page".
      keymap.of([
        { key: "Mod-s", preventDefault: true, run: () => (emit("save"), true) },
        { key: "Mod-=", preventDefault: true, run: () => (emit("zoom", 1), true) },
        { key: "Mod-+", preventDefault: true, run: () => (emit("zoom", 1), true) },
        { key: "Mod--", preventDefault: true, run: () => (emit("zoom", -1), true) },
        { key: "Mod-0", preventDefault: true, run: () => (emit("zoom", 0), true) },
      ]),
      basicSetup,
      html(),
      theme,
      EditorView.lineWrapping,
      EditorView.updateListener.of((update) => {
        if (update.docChanged) emit("update:modelValue", update.state.doc.toString());
      }),
    ],
  });
});

watch(
  () => props.modelValue,
  (value) => {
    if (view && value !== view.state.doc.toString()) {
      view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: value } });
    }
  },
);

// The size comes in through a CSS variable, which CodeMirror can't see change:
// have it re-measure line heights once the new size is applied, or clicks and
// the cursor land on the wrong line.
watch(
  () => props.fontSize,
  () => nextTick(() => view?.requestMeasure()),
);

onBeforeUnmount(() => view?.destroy());

/** Put text at the cursor (replacing any selection); false if the editor
 *  isn't loaded, so the caller can fall back to appending it. */
function insert(text) {
  if (!view) return false;
  view.dispatch(view.state.replaceSelection(text));
  view.focus();
  return true;
}

defineExpose({ insert });
</script>

<template>
  <div ref="host" class="code-editor" :style="{ '--code-font-size': `${fontSize}px` }">
    <p v-if="failed" class="p-2 text-tertiary">Couldn't load the editor.</p>
  </div>
</template>

<style scoped>
.code-editor {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  border: 1px solid var(--color-quaternary);
}
</style>
