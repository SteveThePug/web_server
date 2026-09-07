<script setup>
import Timer from "@/components/util/Timer.vue";
import Time from "@/components/util/Time.vue";
import Radio from "@/components/util/Radio.vue";
import Chat from "@/components/util/Chat.vue";
import CommitHistory from "@/components/util/CommitHistory.vue";

import Intro2 from "./Intro2.vue";
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
</script>

<template>
  <main class="justify-center flex flex-row w-full h-full overflow-x-hidden">
    <div class="outerWrap flex flex-row">
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
        <Intro2 class="intro cell" />
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

/* Tablet: sidebars stack below the sheet */
@media (max-width: 1360px) {
  .outerWrap {
    flex-direction: column;
    align-items: center;
    height: auto;
  }

  .homeGrid {
    order: -1;
    width: 95vw;
    height: 297mm;
    margin-inline: 0;
    box-sizing: border-box;
  }

  .sidebar {
    width: 95vw;
    flex-direction: column;
    align-items: center;
    flex: unset;
    justify-content: space-around;
  }

  .commits-sidebar,
  .steam-sidebar,
  .bookmarks-sidebar {
    width: 100%;
    max-height: 300px;
  }

  .chat-sidebar {
    width: 100%;
    min-height: 400px;
    max-height: 800px;
    height: 25vh;
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
    height: 150vh;
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
