<script setup>
/**
 * STATUS: no route points here — /shrines/<name> is a canvas page now
 * (views/canvas/CanvasPage.vue), created from the /shrines index.
 *
 * Was route `/shrines/gto` — placeholder, renders <Wip>.
 */
import Wip from "@/components/util/Wip.vue";
</script>

<template>
  <main class="items-center flex flex-col">
    <div
      class="page-a4 items-center bdr-1 flex flex-col relative overflow-scroll"
    >
      <Wip />
    </div>
  </main>
</template>
