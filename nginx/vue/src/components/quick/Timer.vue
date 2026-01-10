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
    <div class="flex-col">
        <h4 class="center-content">Timer</h4>
        <!-- Min input and Second input-->
        <div v-if="finished && paused" class="flex-row">
            <input v-model="minutesInput" type="number" min="0" max="59" />
            <input v-model="secondsInput" type="number" min="0" max="59" />
        </div>
        <div v-if="finished && !paused" class="flex-col">
            <h1>Timer finished!</h1>
        </div>
        <div v-if="!finished && paused">
            <h1>Paused</h1>
        </div>
        <div v-if="!finished && !paused" class="flex-col">
            <h1>
                {{ minutes.toString().padStart(2, "0") }}:{{
                    seconds.toString().padStart(2, "0")
                }}
            </h1>
            <h1>
                {{ minutesInput.toString().padStart(2, "0") }}:{{
                    secondsInput.toString().padStart(2, "0")
                }}
            </h1>
        </div>
        <div class="flex-col">
            <button v-if="paused" @click="startTimer">Proceed</button>
            <button v-if="!finished && !paused" @click="pauseTimer">
                Pause
            </button>
            <button v-if="finished ^ paused" @click="resetTimer">Reset</button>
        </div>
    </div>
</template>
