<template>
    <div class="container">
        <img src="/img/borders/utena.png" class="flower tl antirotate" />
        <img src="/img/borders/utena.png" class="flower tr rotate" />
        <img src="/img/borders/utena.png" class="flower bl rotate" />
        <img src="/img/borders/utena.png" class="flower br antirotate" />
        <slot />
    </div>
</template>

<style scoped>
.container {
    position: relative;
    margin: 100px;
}

.flower {
    position: absolute;
    width: 150px;
}
.tl {
    top: -80px;
    left: -80px;
    --start: 0deg;
}
.tr {
    top: -80px;
    right: -80px;
    --start: 90deg;
}
.bl {
    bottom: -80px;
    left: -80px;
    --start: 180deg;
}
.br {
    bottom: -80px;
    right: -80px;
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
