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
.cv-layout h1,
.cv-layout h2,
.cv-layout h3,
.cv-layout h4,
.cv-layout p,
.cv-layout small,
.cv-layout code,
.cv-layout ul,
.cv-layout li,
.cv-layout td,
.cv-layout tr,
.cv-layout table {
  color: #111;
}
.cv-layout h1,
.cv-layout h2,
.cv-layout h3,
.cv-layout h4 {
  margin: 0;
}
.cv-layout a {
  color: #111;
  background-color: transparent !important;
  letter-spacing: normal;
}
.cv-layout input,
.cv-layout textarea {
  color: #111;
  background-color: white;
  border: 1px solid #ccc;
  padding: 0;
  width: auto;
}
.cv-layout input::placeholder,
.cv-layout textarea::placeholder {
  color: #999;
  opacity: 1;
}
.cv-layout table {
  border: 0 solid transparent;
}
.cv-layout tr {
  border-color: transparent;
}
.cv-layout th {
  border: none;
  padding: 0;
}
.cv-layout td {
  padding: 0;
}
</style>
