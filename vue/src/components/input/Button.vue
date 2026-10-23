<!--
  The site's standard button. Wraps a native <button>, so all attributes and
  listeners (including @click) fall through to it via Vue's attribute inheritance.
  A disabled button is dimmed and stops reacting to hover and press, so a form
  that is busy saving looks it.
-->
<script setup></script>

<template>
  <button
    class="px-2 text-primary bg-link-bg text-center border cursor-pointer transition-colors duration-150 ease-in-out hover:bg-surface active:scale-95 disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:bg-link-bg disabled:active:scale-100"
  >
    <slot />
  </button>
</template>
