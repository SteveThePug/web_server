<template>
    <div class="overlay">
        <slot />
    </div>
</template>

<style scoped>
.overlay {
    position: absolute;
    inset: 0; /* full width & height of parent */
    pointer-events: none; /* overlay does not block clicks */
}
</style>
