<script setup>
import { ref } from "vue";

cost time = ref("")
cost weekday = ref("")
cost day = ref("")
cost month = ref("")

function updateDateTime() {
  const date = new Date();
  day.value = date.getDate();
  time.value = date.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" })
  weekday.value = date.toLocaleDateString("en-GB", { weekday: "long" });
  month.value = date.toLocaleDateString("en-GB", { month: "long" });
}

updateDateTime();

setInterval(updateDateTime, 60000);
</script>

<template>
    <div class="flex-col">
        <h4>{{ time }}</h4>
        <h4>{{ weekday }} {{ day }}, {{ month }}</h4>
    </div>
</template>
