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
 */
import { onBeforeUnmount, onMounted, ref, watch } from "vue";

const props = defineProps({
  modelValue: { type: String, default: "" },
});
const emit = defineEmits(["update:modelValue", "save"]);

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
      "&": { height: "100%", fontSize: "13px", backgroundColor: "var(--color-surface-deep)", color: "var(--color-muted)" },
      ".cm-content": { caretColor: "var(--color-primary)" },
      ".cm-gutters": { backgroundColor: "var(--color-surface)", color: "var(--color-quaternary)", border: "none" },
      ".cm-activeLine, .cm-activeLineGutter": { backgroundColor: "var(--color-surface-tint)" },
      "&.cm-focused .cm-selectionBackground, .cm-selectionBackground": { backgroundColor: "#24504a" },
      ".cm-scroller": { fontFamily: "ui-monospace, monospace" },
    },
    { dark: true },
  );

  view = new EditorView({
    parent: host.value,
    doc: props.modelValue,
    extensions: [
      // Before basicSetup so Mod-s wins over the browser's "save page".
      keymap.of([{ key: "Mod-s", preventDefault: true, run: () => (emit("save"), true) }]),
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

onBeforeUnmount(() => view?.destroy());
</script>

<template>
  <div ref="host" class="code-editor">
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
