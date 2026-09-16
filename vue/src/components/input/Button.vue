<!--
  The site's standard button. Wraps a native <button>, so all attributes and
  listeners (including @click) fall through to it via Vue's attribute inheritance.
-->
<script setup></script>

<template>
  <button
    class="text-primary bg-link-bg text-center border cursor-pointer transition-colors duration-150 ease-in-out hover:bg-surface active:scale-95"
  >
    <slot />
  </button>
</template>
