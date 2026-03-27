<script setup>
import { ref, shallowRef, defineAsyncComponent } from "vue";
import CVGeneral from "./CVGeneral.vue";
import CVBackend from "./CVBackend.vue";
import CVFrontend from "./CVFrontend.vue";

const CVHospitality = defineAsyncComponent(() =>
    import("./CVHospitality.vue"),
);

const templates = [
    { label: "General", component: CVGeneral },
    { label: "Backend", component: CVBackend },
    { label: "Frontend", component: CVFrontend },
    { label: "Hospitality", component: CVHospitality },
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
        </div>
        <Transition name="cv-fade" mode="out-in">
            <component :is="currentComponent" :key="selected" />
        </Transition>
    </div>
</template>

<style scoped>
.cv-root {
    background: white;
}

.cv-selector {
    position: sticky;
    top: 0;
    z-index: 40;
    display: flex;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.5rem;
    background: white;
    border-bottom: 1px solid #ddd;
}

.cv-btn {
    padding: 0.4rem 1rem;
    border: 1px solid #333;
    border-radius: 4px;
    background: white;
    color: #333;
    cursor: pointer;
    font-size: 0.9rem;
    transition: background 0.15s, color 0.15s;
}

.cv-btn:hover {
    background: #eee;
}

.cv-btn.active {
    background: #333;
    color: white;
}

.cv-print-btn {
    margin-left: 1rem;
}

.cv-fade-enter-active,
.cv-fade-leave-active {
    transition: opacity 0.15s ease, transform 0.15s ease;
}

.cv-fade-enter-from {
    opacity: 0;
    transform: translateY(6px);
}

.cv-fade-leave-to {
    opacity: 0;
    transform: translateY(-6px);
}

@media print {
    .no-print {
        display: none !important;
    }
}
</style>
