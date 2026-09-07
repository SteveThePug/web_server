<script setup>
import CreateUser from "./CreateUser.vue";
import CreatePost from "./CreatePost.vue";
import CreateFavorite from "./CreateFavorite.vue";
import CreateActivity from "./CreateActivity.vue";
import CreateRowing from "./CreateRowing.vue";
import ManageUsers from "./ManageUsers.vue";
import ManageRadio from "./ManageRadio.vue";
</script>

<template>
  <main class="justify-center flex flex-row w-full h-full">
    <div class="bdr-1 flex flex-col">
      <CreateUser class="bdr-2 bg-surface" />
      <CreatePost class="bdr-2 bg-surface" />
      <CreateFavorite class="bdr-2 bg-surface" />
      <CreateActivity class="bdr-2 bg-surface" />
      <CreateRowing class="bdr-2 bg-surface" />
      <ManageUsers class="bdr-2 bg-surface" />
      <ManageRadio class="bdr-2 bg-surface" />
    </div>
  </main>
</template>
