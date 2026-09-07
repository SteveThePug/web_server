<!--
  Widget title: a heading with a rule underneath.
  The optional #action slot puts a control (e.g. <CreateToggle>) on the right.
-->
<template>
  <div class="w-full border-b border-primary pb-0.5">
    <h1
      class="p-0 m-0"
      :class="{ 'flex items-center justify-between gap-1': $slots.action }"
    >
      <slot />
      <slot name="action" />
    </h1>
  </div>
</template>
