<template>
    <div class="halftone" />
</template>

<style scoped>
div {
    width: 100vw;
    height: 100vh;
    position: sticky;
    left: 0px;
    top: 0px;
    z-index: -1;
}

.halftone {
    --dot_size: 1px;
    --bg_size: 5px;
    --bg_pos: calc(var(--bg_size) / 2);
    --blur: 0%;
    --dot_color: #044;
    --bg_color: #114;

    background-color: var(--bg_color);
    background-image: radial-gradient(
        circle at center,
        var(--dot_color) var(--dot_size),
        transparent var(--blur)
    );
    background-size: var(--bg_size) var(--bg_size);
    background-position: 0 0;

    mask-image: linear-gradient(
        30deg,
        rgba(1, 1, 1, 1) 0%,
        rgba(1, 1, 1, 0.85) 100%
    );
}
</style>
