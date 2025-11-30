<template>
    <!-- Input post display -->
    <div v-if="posting"></div>
    <!-- View post display -->
    <div v-else>
        <div v-for="(post, idx) in posts" :key="post.id || idx">
            <h1>post info</h1>
        </div>
    </div>
</template>

<script>
import { ref, onMounted } from "vue";
export default {
    name: "api-posts",
    setup() {
        const posting = ref(false);
        const posts = ref([]);

        async function fetchRecent() {}

        onMounted(() => {
            fetchRecent();
            setInterval(fetchRecent, 120000);
        });
        return {
            posting,
            posts,
        };
    },
};
</script>
