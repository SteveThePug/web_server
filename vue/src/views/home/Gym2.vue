<script setup>
import { ref, computed, defineAsyncComponent } from "vue";
import Header from "@/components/text/Header.vue";
import { useHomeDataStore } from "@/stores/homeData";
import { useAuthStore } from "@/stores/auth";
import { storeToRefs } from "pinia";

const CreateRowing = defineAsyncComponent(() => import("@/views/admin/CreateRowing.vue"));

const store = useHomeDataStore();
const authStore = useAuthStore();
const { loaded, error, rowingSessions } = storeToRefs(store);
const showCreate = ref(false);

const rows = computed(() => rowingSessions.value.slice().reverse());
const loading = computed(() => !loaded.value);

const metric = ref("distance");
const hovered = ref(null);

const METRICS = [
  { key: "distance", label: "Distance (m)", color: "#55ffbb" },
  { key: "timePer500m", label: "Pace /500m", color: "#ff579a" },
  { key: "calories", label: "Calories", color: "#62ff57" },
];

const activeMetric = computed(() =>
  METRICS.find((m) => m.key === metric.value),
);

// SVG layout constants
const W = 290;
const H = 120;
const PL = 46; // padding left
const PT = 8; // padding top
const PR = 8; // padding right
const PB = 28; // padding bottom
const PLOT_W = W - PL - PR;
const PLOT_H = H - PT - PB;

const chartData = computed(() =>
  rows.value.map((r) => ({
    date: new Date(r.date),
    value: r[metric.value],
    raw: r,
  })),
);

const minVal = computed(() => Math.min(...chartData.value.map((d) => d.value)));
const maxVal = computed(() => Math.max(...chartData.value.map((d) => d.value)));

const points = computed(() => {
  const data = chartData.value;
  const n = data.length;
  if (!n) return [];
  const min = minVal.value;
  const range = maxVal.value - min || 1;
  return data.map((d, i) => ({
    x: PL + (n <= 1 ? PLOT_W / 2 : (i / (n - 1)) * PLOT_W),
    y: PT + PLOT_H - ((d.value - min) / range) * PLOT_H,
    date: d.date,
    value: d.value,
    raw: d.raw,
  }));
});

const polyline = computed(() =>
  points.value.map((p) => `${p.x},${p.y}`).join(" "),
);

const xLabels = computed(() => {
  const data = chartData.value;
  const pts = points.value;
  if (!data.length) return [];
  const indices = new Set([
    0,
    Math.floor((data.length - 1) / 2),
    data.length - 1,
  ]);
  return [...indices].map((i) => ({
    x: pts[i].x,
    label: data[i].date.toLocaleDateString("en-GB", {
      month: "short",
      day: "numeric",
    }),
  }));
});

const yLabels = computed(() => {
  const min = minVal.value;
  const max = maxVal.value;
  return [0, 0.5, 1].map((t) => {
    const raw = Math.round(min + t * (max - min));
    return {
      y: PT + PLOT_H - t * PLOT_H,
      label: metric.value === "timePer500m" ? formatTime(raw) : raw,
    };
  });
});

function formatTime(secs) {
  const m = Math.floor(secs / 60);
  const s = Math.round(secs % 60);
  return `${m}:${String(s).padStart(2, "0")}`;
}

function formatValue(key, val) {
  if (key === "timePer500m") return formatTime(val) + " /500m";
  if (key === "distance") return val + " m";
  if (key === "calories") return Math.round(val) + " kcal";
  return val;
}
</script>

<template>
  <div class="flex flex-col h-full overflow-hidden">
    <Header>
      <span class="flex items-center justify-between w-full">
        {{ showCreate ? "Upload Rowing" : "Rowing" }}
        <button v-if="authStore.user.admin" class="text-sm px-1" @click="showCreate = !showCreate">
          {{ showCreate ? "x" : "+" }}
        </button>
      </span>
    </Header>

    <CreateRowing v-if="showCreate" class="flex-1 p-1" @done="showCreate = false" @cancel="showCreate = false" />
    <div v-else-if="loading" class="flex-1 flex items-center justify-center">
      <p>Loading...</p>
    </div>
    <div v-else-if="error" class="flex-1 flex items-center justify-center">
      <p class="text-tertiary text-xs">{{ error }}</p>
    </div>
    <div v-else class="flex flex-col flex-1 px-1 pb-1 gap-1 overflow-hidden">
      <!-- Metric tabs -->
      <div class="flex gap-1 pt-1">
        <button
          v-for="m in METRICS"
          :key="m.key"
          class="metric-btn text-xs px-2 py-0.5 font-heading border"
          :style="{
            borderColor: m.color,
            color: metric === m.key ? '#1b110e' : m.color,
            backgroundColor: metric === m.key ? m.color : 'transparent',
          }"
          @click="metric = m.key"
        >
          {{ m.label }}
        </button>
      </div>

      <!-- SVG Chart -->
      <div class="flex-1 relative">
        <svg
          :viewBox="`0 0 ${W} ${H}`"
          width="100%"
          height="100%"
          preserveAspectRatio="none"
          class="overflow-visible"
        >
          <!-- Grid lines -->
          <line
            v-for="yl in yLabels"
            :key="yl.y"
            :x1="PL"
            :y1="yl.y"
            :x2="W - PR"
            :y2="yl.y"
            stroke="var(--quaternary)"
            stroke-width="0.5"
          />

          <!-- Area fill -->
          <polygon
            v-if="points.length"
            :points="`${PL},${PT + PLOT_H} ${polyline} ${W - PR},${PT + PLOT_H}`"
            :fill="activeMetric.color"
            fill-opacity="0.08"
          />

          <!-- Line -->
          <polyline
            v-if="points.length"
            :points="polyline"
            :stroke="activeMetric.color"
            stroke-width="1.5"
            fill="none"
            stroke-linejoin="round"
            stroke-linecap="round"
          />

          <!-- Data points -->
          <circle
            v-for="(p, i) in points"
            :key="i"
            :cx="p.x"
            :cy="p.y"
            :r="hovered === i ? 4 : 2"
            :fill="activeMetric.color"
            style="cursor: pointer"
            @mouseenter="hovered = i"
            @mouseleave="hovered = null"
          />

          <!-- Y axis labels -->
          <text
            v-for="yl in yLabels"
            :key="`y${yl.y}`"
            :x="PL - 3"
            :y="yl.y + 3"
            text-anchor="end"
            font-size="10"
            fill="var(--primary)"
            font-family="var(--font_heading)"
          >
            {{ yl.label }}
          </text>

          <!-- X axis labels -->
          <text
            v-for="xl in xLabels"
            :key="`x${xl.x}`"
            :x="xl.x"
            :y="H - 4"
            text-anchor="middle"
            font-size="10"
            fill="var(--primary)"
            font-family="var(--font_heading)"
          >
            {{ xl.label }}
          </text>

          <!-- Axes -->
          <line
            :x1="PL"
            :y1="PT"
            :x2="PL"
            :y2="PT + PLOT_H"
            stroke="var(--primary)"
            stroke-width="0.5"
          />
          <line
            :x1="PL"
            :y1="PT + PLOT_H"
            :x2="W - PR"
            :y2="PT + PLOT_H"
            stroke="var(--primary)"
            stroke-width="0.5"
          />

          <!-- Tooltip -->
          <g v-if="hovered !== null && points[hovered]">
            <rect
              :x="Math.min(points[hovered].x + 4, W - 85)"
              :y="points[hovered].y - 20"
              width="82"
              height="32"
              fill="var(--bg_primary)"
              :stroke="activeMetric.color"
              stroke-width="0.5"
              rx="1"
            />
            <text
              :x="Math.min(points[hovered].x + 7, W - 82)"
              :y="points[hovered].y - 6"
              font-size="12"
              fill="var(--secondary)"
              font-family="var(--font_heading)"
            >
              {{
                points[hovered].date.toLocaleDateString("en-GB", {
                  day: "numeric",
                  month: "short",
                  year: "2-digit",
                })
              }}
            </text>
            <text
              :x="Math.min(points[hovered].x + 7, W - 82)"
              :y="points[hovered].y + 8"
              font-size="14"
              :fill="activeMetric.color"
              font-family="var(--font_heading)"
            >
              {{ formatValue(metric, points[hovered].value) }}
            </text>
          </g>
        </svg>
      </div>

      <!-- Summary stats -->
      <div class="flex justify-between text-xs border-t border-quaternary pt-1">
        <div class="flex flex-col items-center">
          <span class="text-primary font-heading">{{ rows.length }}</span>
          <span class="text-quaternary" style="font-size: 0.6rem"
            >sessions</span
          >
        </div>
        <div class="flex flex-col items-center">
          <span class="text-primary font-heading"
            >{{
              rows.reduce((s, r) => s + r.distance, 0).toLocaleString()
            }}m</span
          >
          <span class="text-quaternary" style="font-size: 0.6rem"
            >total dist</span
          >
        </div>
        <div class="flex flex-col items-center">
          <span class="text-primary font-heading">{{
            formatTime(
              rows.reduce((s, r) => s + r.timePer500m, 0) / (rows.length || 1),
            )
          }}</span>
          <span class="text-quaternary" style="font-size: 0.6rem"
            >avg pace</span
          >
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.metric-btn {
  cursor: pointer;
  transition:
    background-color 0.15s,
    color 0.15s;
  letter-spacing: 0.03em;
}
</style>
