<script setup>
import { ref } from "vue";

const timer = ref(null);

const finished = ref(true);
const paused = ref(true);

const minutesInput = ref(0);
const secondsInput = ref(0);

const minutes = ref(0);
const seconds = ref(0);

const audio = new Audio("/sound/auughhh.mp3");

function tick() {
    seconds.value++;
    if (seconds.value === 60) {
        minutes.value++;
        seconds.value = 0;
    }

    if (minutes.value >= minutesInput.value) {
        if (seconds.value >= secondsInput.value) {
            finished.value = true;
            playFinishedSound();
            clearInterval(timer.value);
        }
    }
}

function startTimer() {
    finished.value = false;
    paused.value = false;
    timer.value = setInterval(tick, 1000);
}

function pauseTimer() {
    if (finished.value) return;

    if (paused.value) {
        timer.value = setInterval(tick, 1000);
        paused.value = false;
    } else {
        clearInterval(timer.value);
        paused.value = true;
    }
}

function resetTimer() {
    finished.value = true;
    paused.value = true;
    clearInterval(timer.value);
    minutes.value = 0;
    seconds.value = 0;
}

function playFinishedSound() {
    audio.play();
}
</script>

<template>
    <div
        class="flex flex-col gap-1 items-center border-primary border p-1 bg-bg_primary"
    >
        <h2 class="items-center">Timer</h2>
        <div v-if="finished && paused" class="flex flex-col">
            <div class="flex flex-row p-2">
                <input
                    class="w-full"
                    v-model="minutesInput"
                    type="range"
                    min="0"
                    max="59"
                />
                <p>{{ minutesInput }}m</p>
            </div>
            <div class="flex flex-row p-2">
                <input
                    class="w-full"
                    v-model="secondsInput"
                    type="range"
                    min="0"
                    max="59"
                />
                <p>{{ secondsInput }}s</p>
            </div>
            <button @click="startTimer">Proceed</button>
        </div>
        <div v-if="finished && !paused" class="flex flex-col">
            <h1>Timer finished!</h1>
            <button @click="resetTimer">Reset</button>
        </div>
        <div v-if="!finished && paused" class="flex flex-col">
            <h1>Paused</h1>
            <button @click="resetTimer">Reset</button>
        </div>
        <div v-if="!finished && !paused" class="flex flex-col">
            <p>
                {{ minutes.toString().padStart(2, "0") }}:{{
                    seconds.toString().padStart(2, "0")
                }}
            </p>
            <p>
                {{ minutesInput.toString().padStart(2, "0") }}:{{
                    secondsInput.toString().padStart(2, "0")
                }}
            </p>
            <button @click="pauseTimer">Pause</button>
        </div>
    </div>
</template>
