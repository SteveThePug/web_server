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
            <CreateUser class="bdr-2 bg-bg_primary" />
            <CreatePost class="bdr-2 bg-bg_primary" />
            <CreateFavorite class="bdr-2 bg-bg_primary" />
            <CreateActivity class="bdr-2 bg-bg_primary" />
            <CreateRowing class="bdr-2 bg-bg_primary" />
            <ManageUsers class="bdr-2 bg-bg_primary" />
            <ManageRadio class="bdr-2 bg-bg_primary" />
        </div>
    </main>
</template>
