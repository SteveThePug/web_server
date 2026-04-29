<script setup>
import Timer from "@/components/util/Timer.vue";
import Time from "@/components/util/Time.vue";
import Radio from "@/components/util/Radio.vue";
import Elle from "@/components/elle/Elle.vue";
import Chat from "@/components/util/Chat.vue";
import MusicPlayer from "@/components/util/MusicPlayer.vue";
import CommitHistory from "@/components/util/CommitHistory.vue";

import Intro from "./Intro.vue";
import Intro2 from "./Intro2.vue";
import BadApple from "./BadApple.vue";
import Miku from "./Miku.vue";
import Stamps from "./Stamps.vue";
import Listening from "./Listening.vue";
import Links from "./Links.vue";
import Feed from "./Feed.vue";
import Collage from "./Collage.vue";
import Favorites from "./Favorites.vue";
// import Gym from "./Gym.vue";
import Gym2 from "./Gym2.vue";
import Consumption from "./Consumption.vue";
import Steam from "./Steam.vue";
import Bookmarks from "./Bookmarks.vue";
</script>

<template>
  <main class="justify-center flex flex-row w-full h-full overflow-x-hidden">
    <div class="outerWrap flex flex-row">
      <div class="sidebar">
        <Time class="time-sidebar sidebar-cell" />
        <Timer class="timer-sidebar sidebar-cell" />
        <Radio class="radio-sidebar sidebar-cell" />
        <CommitHistory class="commits-sidebar flex-1 sidebar-cell" />

        <!-- <Elle class="flex-1" /> -->
        <!-- <MusicPlayer /> -->
        <img
          src="/img/memes/fire-woman.gif"
          alt=""
          width="178"
          height="178"
          class="border-tertiary border-2 sidebar-image box-border w-full bg-tertiary"
          loading="lazy"
        />
      </div>
      <div class="a4page-portrait homeGrid relative bdr-1 border-quaternary">
        <!-- <Intro class="intro" /> -->
        <Intro2 class="intro grid-cell" />
        <!-- <BadApple class="intro" /> -->
        <Listening class="listening grid-cell" />
        <Stamps class="stamps grid-cell" />
        <Feed class="feed grid-cell" />
        <Links class="links grid-cell" />
        <Collage class="collage grid-cell" />
        <Consumption class="consumption grid-cell" />
        <Favorites class="favorites grid-cell" />
        <!-- <Gym class="gym" /> -->
        <Gym2 class="gym grid-cell" />
      </div>
      <div class="sidebar">
        <Steam class="steam-sidebar sidebar-cell" />
        <Bookmarks class="bookmarks-sidebar sidebar-cell" />
        <Chat class="chat-sidebar flex-1 min-h-0 chat-home sidebar-cell" />
        <Miku
          class="sidebar-image miku-image box-border border-tertiary border-2 bg-bg_primary"
        />
      </div>
    </div>
  </main>
</template>

<style scoped>
.grid-cell {
  background-color: var(--bg_primary);
  border-width: 2px;
  border-style: solid;
  border-color: var(--quaternary);
}

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

.sidebar-cell {
  background-color: var(--bg_primary);
  border-width: 2px;
  border-style: solid;
  border-color: var(--quaternary);
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

  /* .sidebar-image { */
  /*     max-height: 150px; */
  /*     max-width: 150px; */
  /*     object-fit: contain; */
  /* } */
}

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

.bg-random {
  background-color: var(--bg_primary);
  background-image: url("/img/miku/miku2.gif");
  background-size: 10px 10px;
}
</style>
