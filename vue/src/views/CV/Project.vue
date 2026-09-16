<!--
  Layout wrapper used inside every CV template: a left gutter (#left slot, e.g.
  dates), a heading row (#top slot) and the default slot for the body.
-->
<script setup></script>

<template>
    <div class="flex flex-col sm:flex-row">
        <div class="sm:w-2/7 mt-auto mb-auto ml-0">
            <slot name="left" />
        </div>
        <div class="w-full p-2">
            <div
                class="flex-row flex place-content-between m-auto place-items-center"
            >
                <slot name="top" />
            </div>
            <slot />
        </div>
    </div>
</template>
