<script setup>
import {
    gbp,
    cycleText,
    cycleTitle,
    showFaster,
    PENDING_ROUTE,
} from "./hotelsFormat.js";

// Table body rows for a list of hotels (already sorted). Columns:
// #, Total, Stay, Travel, Mins, Cycle, Hotel, Route.
defineProps({
    rows: { type: Array, required: true },
    highlightFirst: { type: Boolean, default: true },
});
</script>

<template>
    <template v-for="(r, i) in rows" :key="r.code">
        <tr
            :class="{
                best: highlightFirst && i === 0,
                unpriced: r.total == null,
            }"
        >
            <td class="num">{{ i + 1 }}</td>
            <td class="num total">{{ gbp(r.total) }}</td>
            <td class="num stay">{{ gbp(r.room_price) }}</td>
            <td class="num">{{ gbp(r.transport_total) }}</td>
            <td class="num">{{ r.travel_minutes ?? "" }}</td>
            <td class="num cycle" :title="cycleTitle(r)">{{ cycleText(r) }}</td>
            <td>
                <a
                    :href="r.url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="hotelLink"
                    >{{ r.name }}</a
                >
                <span v-if="r.low_availability" class="lowTag">low availability</span>
            </td>
            <td class="route" :class="{ pending: r.route === PENDING_ROUTE }">
                {{ r.route }}
            </td>
        </tr>
        <tr v-if="showFaster(r)" class="fasterRow">
            <td colspan="7"></td>
            <td class="route">
                faster: {{ r.fastest_minutes }} min for {{ gbp(r.fastest_total) }}
                return via {{ r.fastest_route }}
            </td>
        </tr>
    </template>
</template>
