<template>
    <img src="/img/borders/utena.png" class="flower tl antirotate" />
    <img src="/img/borders/utena.png" class="flower tr rotate" />
    <img src="/img/borders/utena.png" class="flower bl rotate" />
    <img src="/img/borders/utena.png" class="flower br antirotate" />
    <slot />
</template>

<style scoped>
.flower {
    position: absolute;
    width: 150px;
}
.tl {
    top: 100px;
    left: 0px;
    --start: 0deg;
}
.tr {
    top: 100px;
    right: 0;
    --start: 90deg;
}
.bl {
    bottom: -100px;
    left: 0;
    --start: 180deg;
}
.br {
    bottom: -100px;
    right: 0;
    --start: 270deg;
}
.rotate {
    animation: spin 3s linear infinite;
}

.antirotate {
    animation: spin 3s linear infinite reverse;
}

@keyframes spin {
    from {
        transform: rotate(var(--start));
    }
    to {
        transform: rotate(calc(var(--start) + 360deg));
    }
}
</style>
