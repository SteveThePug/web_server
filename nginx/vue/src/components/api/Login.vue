<template>
    <div v-if="loggedIn">
        <h1>login</h1>
    </div>
</template>

<script>
import { ref, onMounted } from "vue";
export default {
    name: "api-login",
    setup() {
        const loggedIn = ref(false);
        onMounted(() => {
            loggedIn.value = true;
        });
        return { loggedIn };
    },
};
</script>
