<script setup>
/**
 * Route `/stp` — the personal home page: an A4-shaped widget grid flanked by two
 * sidebars.
 *
 * Layout is entirely CSS `grid-template-areas`: each child gets a `grid-area`
 * class and the three breakpoints below (desktop / <=1360px tablet / <=700px
 * phone) re-draw the whole area map rather than moving elements around.
 *
 * Below 1360px there is no room for the sidebars, so their widgets collapse
 * into a single tabbed panel under the sheet. That is done purely in CSS — the
 * tab bar only sets `data-tab` on the wrapper and the media query hides the
 * other widgets — so every widget stays mounted exactly once at every width
 * (Chat holds a WebSocket, Steam a rotation timer) and resizing across the
 * breakpoint loses no state.
 *
 * The widgets do not fetch individually — they all read stores/homeData.js, which
 * issues one GraphQL query for the whole page.
 *
 * Note which variants are wired up: Intro3 and Gym2, not Intro/Intro2/Gym.
 */
import Timer from "@/components/util/Timer.vue";
import Time from "@/components/util/Time.vue";
import Radio from "@/components/util/Radio.vue";
import Chat from "@/components/util/Chat.vue";
import CommitHistory from "@/components/util/CommitHistory.vue";

import Intro3 from "./Intro3.vue";
import Miku from "./Miku.vue";
import Stamps from "./Stamps.vue";
import Listening from "./Listening.vue";
import Links from "./Links.vue";
import Feed from "./Feed.vue";
import Collage from "./Collage.vue";
import Favorites from "./Favorites.vue";
import Gym2 from "./Gym2.vue";
import Consumption from "./Consumption.vue";
import Steam from "./Steam.vue";
import Bookmarks from "./Bookmarks.vue";

import { ref } from "vue";

// Small-screen tab bar. `id` matches the [data-tab] selectors in the tablet
// media query below.
const TABS = [
  { id: "chat", label: "Chat" },
  { id: "bookmarks", label: "Bookmarks" },
  { id: "commits", label: "Commits" },
  { id: "steam", label: "Steam" },
];
const activeTab = ref("chat");
</script>

<template>
  <main class="justify-center flex flex-row w-full h-full overflow-x-hidden">
    <div class="outerWrap flex flex-row" :data-tab="activeTab">
      <nav class="widget-tabs" aria-label="Widgets">
        <button
          v-for="tab in TABS"
          :key="tab.id"
          type="button"
          class="widget-tab"
          :class="{ 'is-active': activeTab === tab.id }"
          :aria-pressed="activeTab === tab.id"
          @click="activeTab = tab.id"
        >
          {{ tab.label }}
        </button>
      </nav>
      <div class="sidebar">
        <Time class="time-sidebar cell" />
        <Timer class="timer-sidebar cell" />
        <Radio class="radio-sidebar cell" />
        <CommitHistory class="commits-sidebar flex-1 cell" />
        <img
          src="/img/memes/fire-woman.gif"
          alt=""
          width="178"
          height="178"
          class="border-tertiary border-2 sidebar-image box-border w-full bg-tertiary"
          loading="lazy"
        />
      </div>
      <div class="page-a4 homeGrid relative bdr-1 border-quaternary">
        <Intro3 class="intro cell" />
        <Listening class="listening cell" />
        <Stamps class="stamps cell" />
        <Feed class="feed cell" />
        <Links class="links cell" />
        <Collage class="collage cell" />
        <Consumption class="consumption cell" />
        <Favorites class="favorites cell" />
        <Gym2 class="gym cell" />
      </div>
      <div class="sidebar">
        <Steam class="steam-sidebar cell" />
        <Bookmarks class="bookmarks-sidebar cell" />
        <Chat class="chat-sidebar flex-1 min-h-0 chat-home cell" />
        <Miku
          class="sidebar-image miku-image box-border border-tertiary border-2 bg-surface"
        />
      </div>
    </div>
  </main>
</template>

<style scoped>
/* Every widget, in the grid or a sidebar */
.cell {
  background-color: var(--color-surface);
  border: 2px solid var(--color-quaternary);
}

/* Grid placement — see grid-template-areas in .homeGrid below */
.intro {
  grid-area: intro;
}
.listening {
  grid-area: listening;
}
.stamps {
  grid-area: stamps;
}
.feed {
  grid-area: feed;
}
.links {
  grid-area: links;
}
.collage {
  grid-area: collage;
}
.consumption {
  grid-area: consumption;
}
.gym {
  grid-area: gym;
}
.favorites {
  grid-area: favorites;
}

.sidebar {
  margin: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  width: 15rem;
  min-height: 0;
  gap: 5px;
}

.outerWrap {
  height: 310mm;
  gap: 10px;
}

/* Desktop: 10x10 grid on the A4 sheet */
.homeGrid {
  display: grid;
  gap: 5px;
  grid-template-columns: repeat(10, 1fr);
  grid-template-rows: repeat(10, 1fr);
  grid-template-areas:
    "intro       intro       intro       intro       intro       intro       listening   listening   listening   listening"
    "intro       intro       intro       intro       intro       intro       listening   listening   listening   listening"
    "intro       intro       intro       intro       intro       intro       listening   listening   listening   listening"
    "intro       intro       intro       intro       intro       intro       stamps      stamps      stamps      stamps"
    "feed        feed        feed        links       links       collage     collage     collage     collage     collage"
    "feed        feed        feed        links       links       collage     collage     collage     collage     collage"
    "feed        feed        feed        links       links       collage     collage     collage     collage     collage"
    "feed        feed        feed        links       links       collage     collage     collage     collage     collage"
    "consumption consumption consumption consumption gym         gym         gym         favorites   favorites   favorites"
    "consumption consumption consumption consumption gym         gym         gym         favorites   favorites   favorites";
}

.chat-home {
  max-height: 800px;
}

.miku-image {
  height: 15rem;
}

/* Small-screen only; switched on in the tablet media query */
.widget-tabs {
  display: none;
}

.widget-tab {
  flex: 1;
  min-width: 0;
  padding: 0.5rem 0.25rem;
  color: var(--color-primary);
  background-color: var(--color-link-bg);
  border: 2px solid var(--color-quaternary);
  font-family: var(--font-heading);
  font-size: 1.125rem;
  letter-spacing: 0.025em;
  cursor: pointer;
  transition:
    background-color 120ms ease,
    border-color 120ms ease,
    color 120ms ease;
}
.widget-tab:hover {
  border-color: var(--color-primary);
}
.widget-tab.is-active {
  border-color: var(--color-primary);
  color: var(--color-tertiary);
  background-color: var(--color-surface-tint);
}

/* Tablet: the sidebars collapse into one tabbed panel below the sheet */
@media (max-width: 1360px) {
  .outerWrap {
    flex-direction: column;
    align-items: center;
    height: auto;
    gap: 5px;
    padding-bottom: 10px;
  }

  /* Visual order is sheet, tab bar, selected widget — whatever the DOM order */
  .homeGrid {
    order: -2;
    width: 95vw;
    height: 297mm;
    margin-inline: 0;
    box-sizing: border-box;
  }

  .widget-tabs {
    order: -1;
    display: flex;
    gap: 5px;
    width: 95vw;
    margin-top: 5px;
  }

  /* The two sidebars stop being boxes, so their widgets become direct flex
     children of .outerWrap and the one selected widget sits under the tabs. */
  .sidebar {
    display: contents;
  }

  /* One fixed panel size for every tab, so switching doesn't make the page
     jump. The .outerWrap prefix is for specificity: it has to beat the
     widgets' own scoped height rules (Steam's 54mm, Chat's 100%). */
  .outerWrap .commits-sidebar,
  .outerWrap .steam-sidebar,
  .outerWrap .bookmarks-sidebar,
  .outerWrap .chat-sidebar {
    flex: none;
    width: 95vw;
    height: clamp(360px, 70vh, 640px);
    max-height: none;
  }

  .outerWrap:not([data-tab="commits"]) .commits-sidebar,
  .outerWrap:not([data-tab="steam"]) .steam-sidebar,
  .outerWrap:not([data-tab="bookmarks"]) .bookmarks-sidebar,
  .outerWrap:not([data-tab="chat"]) .chat-sidebar {
    display: none;
  }

  .time-sidebar,
  .sidebar-image,
  .radio-sidebar,
  .timer-sidebar {
    display: none;
  }
}

/* Phone: 3-column grid, collage and gym hidden */
@media (max-width: 700px) {
  .homeGrid {
    border-image: none;
    border-width: 0;
    grid-template-columns: 1fr 1fr 1fr;
    grid-template-rows: repeat(11, 1fr);
    /* The floor keeps the 11 rows usable on short phones, where 150vh alone
       leaves the two-row Listening cell too small to show its album art. */
    height: max(150vh, 1000px);
    grid-template-areas:
      "intro       intro       intro"
      "intro       intro       intro"
      "listening   stamps      stamps"
      "listening   feed        feed"
      "links   feed        feed"
      "links       feed        feed"
      "links       feed        feed"
      "favorites       feed        feed"
      "favorites   consumption consumption"
      "favorites   consumption consumption"
      "favorites   consumption consumption";
  }

  .collage {
    display: none;
  }

  .gym {
    display: none;
  }
}
</style>
