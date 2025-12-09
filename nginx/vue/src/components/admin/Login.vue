<script setup>
import { ref, onMounted } from "vue";

const loggedIn = ref(false);
</script>

<template>
    <div v-if="!loggedIn">
        <h1>login</h1>
        <textarea>Username</textarea>
        <textarea>Password</textarea>
    </div>
</template>
