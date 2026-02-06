<script setup>
import { ref, onMounted, onBeforeUnmount, watchEffect } from "vue";

const container = ref(null);
const direction = ref(1); // 1 = down, -1 = up
const isPaused = ref(false);
const animationFrameId = ref(null);
const lastTime = ref(0);

const SPEED = 0.03; // px per ms (better for frame rate independence)
const PAUSE_DURATION = 2000; // ms at top/bottom

// Use requestAnimationFrame for smoother scrolling
function scrollStep(timestamp) {
    if (isPaused.value) return;

    if (!lastTime.value) lastTime.value = timestamp;
    const deltaTime = timestamp - lastTime.value;
    lastTime.value = timestamp;

    if (!container.value) return;

    const el = container.value;
    const scrollDistance = SPEED * deltaTime * direction.value;

    el.scrollTop += scrollDistance;

    const reachedBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 1;
    const reachedTop = el.scrollTop <= 1;

    if (reachedBottom || reachedTop) {
        direction.value *= -1;
        isPaused.value = true;

        setTimeout(() => {
            isPaused.value = false;
            if (!isPaused.value) {
                animationFrameId.value = requestAnimationFrame(scrollStep);
            }
        }, PAUSE_DURATION);

        return;
    }

    animationFrameId.value = requestAnimationFrame(scrollStep);
}

function pauseScroll() {
    isPaused.value = true;
    lastTime.value = 0; // Reset time tracking
}

function resumeScroll() {
    if (isPaused.value) {
        isPaused.value = false;
        animationFrameId.value = requestAnimationFrame(scrollStep);
    }
}

function startScrolling() {
    if (!isPaused.value) {
        animationFrameId.value = requestAnimationFrame(scrollStep);
    }
}

function stopScrolling() {
    if (animationFrameId.value) {
        cancelAnimationFrame(animationFrameId.value);
        animationFrameId.value = null;
    }
}

onMounted(() => {
    startScrolling();
});

onBeforeUnmount(() => {
    stopScrolling();
});

// Optional: Auto-pause when component is not visible
const observer = ref(null);
onMounted(() => {
    if ("IntersectionObserver" in window) {
        observer.value = new IntersectionObserver(
            (entries) => {
                isPaused.value = !entries[0].isIntersecting;
                if (!isPaused.value) {
                    startScrolling();
                }
            },
            { threshold: 0.1 },
        );

        if (container.value) {
            observer.value.observe(container.value);
        }
    }
});

onBeforeUnmount(() => {
    if (observer.value) {
        observer.value.disconnect();
    }
});
</script>

<template>
    <div
        ref="container"
        @mouseover="pauseScroll"
        @mouseleave="resumeScroll"
        class="overflow-y-auto"
    >
        <slot />
    </div>
</template>
