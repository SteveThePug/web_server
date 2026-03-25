<script setup>
import { ref } from "vue";

const props = defineProps({
    modelValue: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["update:modelValue"]);

function toggle() {
    emit("update:modelValue", !props.modelValue);
}
</script>
<template>
    <button
        @click="toggle"
        class="box-content border-2 border-primary w-20 h-fit rounded-full cursor-pointer"
        :class="[props.modelValue ? 'bg-bg_secondary' : 'bg-bg_primary']"
    >
        <svg
            viewBox="0 0 40 40"
            class="w-10 h-10 transition-all duration-300 ease-in-out"
            :class="[props.modelValue ? 'ml-10' : 'ml-0']"
        >
            <circle class="fill-primary" cx="20" cy="20" r="20" />
        </svg>
    </button>
</template>
