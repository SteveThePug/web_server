<template>
    <div class="overlay">
        <slot />
    </div>
</template>

<style scoped>
.overlay {
    position: fixed; /* not absolute */
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    pointer-events: none;
}
</style>
