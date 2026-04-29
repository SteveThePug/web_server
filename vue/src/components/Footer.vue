<script setup>
import { ref, onMounted, onUnmounted } from "vue";

const clock = ref("");
let timer;

function updateClock() {
  const now = new Date();
  clock.value = now.toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

onMounted(() => {
  updateClock();
  timer = setInterval(updateClock, 1000);
});

onUnmounted(() => {
  clearInterval(timer);
});

const user = "visitor";
</script>

<template>
  <footer class="waybar">
    <div class="modules-left">
      <span class="workspace active">ツ</span>
    </div>

    <div class="modules-right">
      <span class="module greeting">Hi, {{ user }}!</span>
      <span class="module cpu hide-sm">CPU 3%</span>
      <span class="module mem hide-sm">MEM 42%</span>
      <span class="module disk hide-sm">DISK 67%</span>
      <span class="module network hide-sm">↑ 12K ↓ 84K</span>
      <span class="module battery hide-sm">BAT 98%</span>
      <span class="module clock">{{ clock }}</span>
    </div>
  </footer>
</template>

<style scoped>
.waybar {
  font-family: "URWGothic-Book", monospace;
  background-color: var(--bg_primary);
  color: var(--primary);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 8px;
  font-size: 14px;
  min-height: 36px;
  flex-shrink: 0;
}

.modules-left {
  display: flex;
  gap: 2px;
}

.workspace {
  background: var(--quaternary);
  border: none;
  border-bottom: 2px solid var(--secondary);
  color: var(--secondary);
  padding: 2px 10px;
  font-family: inherit;
  font-size: 14px;
}

.modules-right {
  display: flex;
  align-items: center;
}

.module {
  padding: 2px 12px;
  border-left: 1px solid var(--tertiary);
}

.module:first-child {
  border-left: none;
}

.greeting {
  color: var(--secondary);
}

.clock {
  color: var(--tertiary);
}

.cpu,
.mem,
.disk {
  color: var(--primary);
}

.network {
  color: var(--secondary);
}

.battery {
  color: var(--primary);
}

@media (max-width: 800px) {
  .waybar {
    font-size: 11px;
    padding: 4px 4px;
  }

  .workspace {
    padding: 2px 6px;
    font-size: 11px;
  }

  .module {
    padding: 2px 6px;
  }

  .hide-sm {
    display: none;
  }
}
</style>
