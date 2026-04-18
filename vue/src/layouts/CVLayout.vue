<script setup>
import { RouterView } from "vue-router";
</script>

<template>
    <div class="cv-layout">
        <RouterView />
    </div>
</template>

<style scoped>
.cv-layout {
    min-height: 100vh;
    background: white;
    color: #111;
}
</style>

<style>
/* Reset global element styles within CV layout */
.cv-layout h1, .cv-layout h2, .cv-layout h3, .cv-layout h4 {
    color: inherit;
    font-family: inherit;
    margin: 0;
}
.cv-layout p, .cv-layout small, .cv-layout code, .cv-layout ul, .cv-layout li {
    color: inherit;
}
.cv-layout a {
    color: inherit;
    background-color: transparent !important;
    font-family: inherit;
    text-align: inherit;
    letter-spacing: normal;
}
.cv-layout input, .cv-layout textarea {
    color: inherit;
    border-color: #ccc;
    border-width: 1px;
    padding: 0;
    width: auto;
}
.cv-layout input::placeholder, .cv-layout textarea::placeholder {
    color: #999;
    opacity: 1;
}
.cv-layout table {
    border-color: transparent;
    border-width: 0;
    color: inherit;
}
.cv-layout tr {
    border-color: inherit;
    color: inherit;
}
.cv-layout th {
    border-right-width: 0;
    border-style: none;
    border-color: inherit;
    padding: 0;
}
.cv-layout td {
    padding: 0;
    color: inherit;
}
.cv-layout textarea {
    color: inherit;
    background-color: white;
}
</style>
