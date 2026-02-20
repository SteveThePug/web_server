<script setup></script>

<template>
    <div class="flex-row flex">
        <div class="w-2/7 p-5 m-auto">
            <slot name="left" />
        </div>
        <div class="w-full p-2">
            <div
                class="flex-row flex place-content-between m-auto place-items-center"
            >
                <slot name="top" />
            </div>
            <slot />
        </div>
    </div>
</template>
