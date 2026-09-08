<script setup>
import { CUSTOM_ORIGIN } from "./useHotelForm.js";

defineProps({
    form: { type: Object, required: true },
});
</script>

<template>
    <label>
        Search area
        <input type="text" v-model="form.location" maxlength="80" required />
    </label>
    <label>
        Rooms &amp; guests
        <input
            type="text"
            v-model="form.guests"
            placeholder='"1", "2+1", "2,2+2"'
            maxlength="40"
        />
        <span class="hint"
            >rooms separated by commas, each ADULTS or ADULTS+CHILDREN</span
        >
    </label>
    <label>
        Railcard holders
        <input
            type="number"
            v-model.number="form.railcardHolders"
            min="0"
            max="8"
        />
        <span class="hint">adults getting 1/3 off off-peak rail</span>
    </label>
    <label>
        Origin station
        <select v-model="form.originKey">
            <option v-for="o in form.origins" :key="o.key" :value="o.key">
                {{ o.name }}
            </option>
            <option v-if="!form.origins.length" value="clapham junction">
                Clapham Junction
            </option>
            <option :value="CUSTOM_ORIGIN">Custom lat,lon…</option>
        </select>
        <input
            v-if="form.originKey === CUSTOM_ORIGIN"
            type="text"
            v-model="form.customOrigin"
            placeholder="51.5031,-0.1132"
            class="mt-1"
            required
        />
    </label>
    <label>
        Leave at
        <input type="time" v-model="form.depart" required />
        <span class="hint">16:00–19:00 Mon–Fri is peak (dearer, no railcard)</span>
    </label>
    <label>
        Max miles from search area
        <input
            type="number"
            v-model="form.maxMiles"
            min="0.5"
            max="50"
            step="0.5"
            placeholder="any"
        />
    </label>
    <label>
        Max travel minutes
        <input
            type="number"
            v-model.number="form.maxTravelMin"
            min="10"
            max="240"
        />
        <span class="hint">slower routes are ignored when picking the cheapest</span>
    </label>
</template>
