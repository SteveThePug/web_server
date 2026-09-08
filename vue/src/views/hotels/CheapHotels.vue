<script setup>
import { ref, computed, onMounted } from "vue";
import Header from "@/components/text/Header.vue";
import Paragraph from "@/components/text/Paragraph.vue";
import InlineLink from "@/components/text/InlineLink.vue";
import SingleStayPanel from "./SingleStayPanel.vue";
import CheapestNightPanel from "./CheapestNightPanel.vue";
import { useHotelForm, loadFrom, saveTo } from "./useHotelForm.js";

const MODE_KEY = "cheap-hotels-mode";

const TABS = [
    { key: "single", label: "Single stay", component: SingleStayPanel },
    { key: "scan", label: "Cheapest night", component: CheapestNightPanel },
];

const saved = loadFrom(MODE_KEY);
const mode = ref(TABS.some((t) => t.key === saved) ? saved : "single");
const panel = computed(() => TABS.find((t) => t.key === mode.value).component);

// One shared form object for both panels, so origin, guests and the like keep
// their values when switching tabs. Loaded before the panels render.
const form = useHotelForm();
form.load();

function setMode(key) {
    mode.value = key;
    saveTo(MODE_KEY, key);
}

onMounted(() => form.loadOrigins());
</script>

<template>
    <main class="hotels flex justify-center px-4 py-10">
        <div class="max-w-5xl w-full flex flex-col gap-6">
            <section>
                <Header>Cheap Travelodge Finder</Header>
                <Paragraph>
                    Ranks the cheapest Travelodges for a stay in London
                    <em>including the cost of getting there and back</em> from
                    a chosen station, so a £45 room forty minutes out can be
                    compared fairly with a £90 room in Zone 1. Room prices come
                    live from Travelodge; fares and routes come from
                    <InlineLink
                        href="https://tfl.gov.uk/plan-a-journey/"
                        target="_blank"
                        >TfL Journey Planner</InlineLink
                    >. Every hotel also gets a cycling time from the origin, for
                    when the bike is the cheapest transport of all.
                    <em>Single stay</em> prices one check-in date;
                    <em>Cheapest night</em> scans a whole date range to find
                    the night worth booking.
                </Paragraph>
            </section>

            <div class="tabBar" role="tablist">
                <button
                    v-for="t in TABS"
                    :key="t.key"
                    type="button"
                    role="tab"
                    class="rankBtn tabBtn"
                    :class="{ active: mode === t.key }"
                    :aria-selected="mode === t.key"
                    @click="setMode(t.key)"
                >
                    {{ t.label }}
                </button>
            </div>

            <KeepAlive>
                <component :is="panel" :key="mode" :form="form" />
            </KeepAlive>
        </div>
    </main>
</template>

<style src="./hotels-shared.css"></style>
