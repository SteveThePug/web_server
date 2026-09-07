<script setup>
import { ref, shallowRef } from "vue";
import { RouterLink } from "vue-router";
// import CVGeneral from "./CVGeneral.vue";
import CVProgramming from "./CVProgramming.vue";
import CoverLetters from "../unused/CoverLetters.vue";
// import CVSecurity from "./CVSecurity.vue";
// import CVBackend from "./CVBackend.vue";
// import CVFrontend from "./CVFrontend.vue";
// import CVTemp from "./CVTemp.vue";
// import CVElectrical from "./CVElectrical.vue";
// import CVHospitality from "./CVHospitality.vue";
// import CVAnalyst from "./CVAnalyst.vue";

const templates = [
    { label: "Programming", component: CVProgramming },
    { label: "Cover Letter", component: CoverLetters },
    // { label: "General", component: CVGeneral },
    // { label: "Security", component: CVSecurity },
    // { label: "Backend", component: CVBackend },
    // { label: "Frontend", component: CVFrontend },
    // { label: "Temp", component: CVTemp },
    // { label: "Electrical", component: CVElectrical },
    // { label: "Hospitality", component: CVHospitality },
    // { label: "Analyst", component: CVAnalyst },
];

const selected = ref(0);
const currentComponent = shallowRef(templates[0].component);

function select(index) {
    selected.value = index;
    currentComponent.value = templates[index].component;
}

function print() {
    window.print();
}
</script>

<template>
    <div class="cv-root">
        <div class="no-print cv-selector">
            <button
                v-for="(t, i) in templates"
                :key="t.label"
                :class="['cv-btn', { active: selected === i }]"
                @click="select(i)"
            >
                {{ t.label }}
            </button>
            <button class="cv-btn cv-print-btn" @click="print()">Print</button>
            <RouterLink to="/cv/jobs" class="cv-btn cv-jobs-btn"
                >Jobs</RouterLink
            >
        </div>
        <Transition name="cv-fade" mode="out-in">
            <component :is="currentComponent" :key="selected" />
        </Transition>
    </div>
</template>

<style scoped>
.cv-root {
    background: var(--color-paper);
}

.cv-selector {
    position: sticky;
    top: 0;
    z-index: 40;
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.5rem;
    background: var(--color-paper);
    border-bottom: 1px solid var(--color-line);
}

.cv-print-btn {
    margin-left: 1rem;
}

.cv-jobs-btn {
    margin-left: auto;
}

.cv-fade-enter-active,
.cv-fade-leave-active {
    transition:
        opacity 0.15s ease,
        transform 0.15s ease;
}

.cv-fade-enter-from {
    opacity: 0;
    transform: translateY(6px);
}

.cv-fade-leave-to {
    opacity: 0;
    transform: translateY(-6px);
}
</style>

<style src="./cv-shared.css"></style>
