<script setup>
import Header from "@/components/text/Header.vue";
import { ref } from "vue";

const time = ref("");
const weekday = ref("");
const day = ref("");
const month = ref("");

function updateDateTime() {
    const date = new Date();
    day.value = date.getDate();
    time.value = date.toLocaleTimeString("en-GB", {
        hour: "2-digit",
        minute: "2-digit",
    });
    weekday.value = date.toLocaleDateString("en-GB", { weekday: "long" });
    month.value = date.toLocaleDateString("en-GB", { month: "long" });
}

updateDateTime();

setInterval(updateDateTime, 60000);
</script>

<template>
    <div class="items-center flex flex-col bg-bg_primary border-primary border">
        <h1>{{ time }}</h1>
        <h1>{{ weekday }} {{ day }}, {{ month }}</h1>
    </div>
</template>
