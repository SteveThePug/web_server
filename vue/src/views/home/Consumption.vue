<script setup>
import AutoScroll from "@/components/util/AutoScroll.vue";
import LinkTable from "@/components/util/LinkTable.vue";
import Header from "@/components/text/Header.vue";

import { ref, defineAsyncComponent } from "vue";
import { useActivityStore } from "@/stores/activity";
import { useAuthStore } from "@/stores/auth";

const CreateActivity = defineAsyncComponent(() => import("@/views/admin/CreateActivity.vue"));

const activityStore = useActivityStore();
const authStore = useAuthStore();
const showCreate = ref(false);
</script>

<template>
    <div class="flex flex-col items-center">
        <Header>
            <span class="flex items-center justify-between w-full">
                {{ showCreate ? "Create Activity" : "Consumption" }}
                <button v-if="authStore.user.admin" class="text-sm px-1" @click="showCreate = !showCreate">
                    {{ showCreate ? "x" : "+" }}
                </button>
            </span>
        </Header>
        <CreateActivity v-if="showCreate" class="flex-1 w-full p-1" @done="showCreate = false" @cancel="showCreate = false" />
        <AutoScroll v-if="!showCreate" class="flex-1 w-full">
            <LinkTable variant="table" class="w-full" :items="activityStore.activity" />
        </AutoScroll>
    </div>
</template>
