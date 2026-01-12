<script setup>
// Get the current date and time, make time reactive
// Get the month name, weekday name and time should be hh:mm:ss
import { ref } from "vue";

const date = new Date();
const day = date.getDate();
const time = ref(
    date.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" }),
);
const weekday = date.toLocaleDateString("en-GB", { weekday: "long" });
const month = date.toLocaleDateString("en-GB", { month: "long" });

setInterval(() => {
    time.value = date.toLocaleTimeString("en-GB", {
        hour: "2-digit",
        minute: "2-digit",
    });
}, 1000);
</script>

<template>
    <div class="flex-col">
        <h4>{{ time }}</h4>
        <h4>{{ weekday }} {{ day }}, {{ month }}</h4>
    </div>
</template>
