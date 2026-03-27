<script setup>
import Headline from "@/components/text/Headline.vue";
import { computed } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();

const parentPath = computed(() => {
    const segments = route.path.split("/").filter(Boolean);
    if (segments.length == 1) {
        return "/";
    } else {
        segments.pop();
        return segments.length ? "/" + segments.join("/") : null;
    }
});

const faces = [
    "^_^",
    "¯\\_(ツ)_/¯",
    "(◕‿◕✿)",
    "ಠ_ಠ",
    "ʘ‿ʘ",
    "^̮^",
    ">_>",
    "¬_¬",
    "˙ ͜ʟ˙",
    "( ͡° ͜ʖ ͡°)",
    "[̲̅$̲̅(̲̅5̲̅)̲̅$̲̅]",
    "(ง'̀-'́)ง",
    "\ (•◡•) /",
    "( ͡ᵔ ͜ʖ ͡ᵔ )",
    "ᕙ(⇀‸↼‶)ᕗ",
    "⚆ _ ⚆",
    "(｡◕‿◕｡)",
    "(╯°□°)╯︵ ʞooqǝɔɐɟ",
    "̿ ̿ ̿'̿'\̵͇̿̿\з=(•_•)=ε/̵͇̿̿/'̿'̿ ̿",
    "(☞ﾟヮﾟ)☞ ☜(ﾟヮﾟ☜)",
];
const faces_string = faces.join("      ");
</script>

<template>
    <nav class="flex flex-row w-full h-fit border border-primary bg-bg_primary">
<RouterLink class="bdr-2 bg-bg_primary" v-if="parentPath" :to="parentPath">
            <span>UP</span>
        </RouterLink>
        <Headline class="border flex-1 max-w-full">
            <code class="whitespace-pre">{{ faces_string }}</code>
        </Headline>
    </nav>
</template>

<style scoped>
.left {
    position: fixed;
    top: 0;
    left: 0;
}
</style>
