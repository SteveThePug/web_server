<script setup>
import { ref } from "vue";
import ToggleButton from "@/components/input/ToggleButton.vue";

const props = defineProps({
    modelValue: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["update:modelValue"]);
const toggleButtonRef = ref(null);

const updateValue = (newValue) => {
    emit("update:modelValue", newValue);
};
const handleClick = () => {
    toggleButtonRef.value?.$el?.click();
};
</script>
<template>
    <div
        class="w-full border-b border-primary cursor-pointer"
        @click="handleClick"
    >
        <h3 class="pl-2 m-0">
            <slot />
        </h3>
        <ToggleButton
            class="pointer-events-none"
            :model-value="props.modelValue"
            @update:model-value="updateValue"
            @click.stop
            ref="toggleButtonRef"
        />
    </div>
</template>
