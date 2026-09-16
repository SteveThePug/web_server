<script setup>
/**
 * Countdown widget with minute/second sliders and a sound on completion.
 *
 * `state` is an explicit state machine — "idle" (sliders), "running",
 * "paused" and "finished" — one screen per state. Completion is decided by
 * comparing total elapsed seconds with the total target, so it does not depend
 * on the minute and second fields crossing their targets independently.
 */
import Button from "@/components/input/Button.vue";
import Header from "@/components/text/Header.vue";

import { ref, onUnmounted } from "vue";

const timer = ref(null);

/** "idle" | "running" | "paused" | "finished" */
const state = ref("idle");

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

  const elapsed = minutes.value * 60 + seconds.value;
  const target = Number(minutesInput.value) * 60 + Number(secondsInput.value);

  if (elapsed >= target) {
    state.value = "finished";
    playFinishedSound();
    clearInterval(timer.value);
  }
}

function startTimer() {
  state.value = "running";
  timer.value = setInterval(tick, 1000);
}

function pauseTimer() {
  if (state.value === "finished") return;

  if (state.value === "paused") {
    timer.value = setInterval(tick, 1000);
    state.value = "running";
  } else {
    clearInterval(timer.value);
    state.value = "paused";
  }
}

/** Back to the slider screen, ready to set a new countdown. */
function resetTimer() {
  state.value = "idle";
  clearInterval(timer.value);
  minutes.value = 0;
  seconds.value = 0;
}

onUnmounted(() => {
  clearInterval(timer.value);
});

function playFinishedSound() {
  audio.play();
}
</script>

<template>
  <div class="timer-root flex flex-col gap-1 p-1 items-center">
    <Header>Timer</Header>
    <div v-if="state === 'idle'" class="flex flex-col">
      <div class="flex flex-row p-2 place-content-around">
        <input
          class="w-2/3"
          v-model="minutesInput"
          type="range"
          min="0"
          max="59"
          aria-label="Minutes"
        />
        <p>{{ minutesInput }}m</p>
      </div>
      <div class="flex flex-row p-2 place-content-around">
        <input
          class="w-2/3"
          v-model="secondsInput"
          type="range"
          min="0"
          max="59"
          aria-label="Seconds"
        />
        <p>{{ secondsInput }}s</p>
      </div>
      <Button @click="startTimer">Proceed</Button>
    </div>
    <div v-if="state === 'finished'" class="flex flex-col">
      <h1>Timer finished!</h1>
      <Button @click="resetTimer">Reset</Button>
    </div>
    <div v-if="state === 'paused'" class="flex flex-col">
      <h1>Paused</h1>
      <Button @click="resetTimer">Reset</Button>
    </div>
    <div v-if="state === 'running'" class="flex flex-col">
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
      <Button @click="pauseTimer">Pause</Button>
    </div>
  </div>
</template>

<style scoped>
@media (max-width: 850px) {
  .timer-root {
    padding: 2px;
    gap: 2px;
  }
}
</style>
