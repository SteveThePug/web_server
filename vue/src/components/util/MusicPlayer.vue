<script setup>
import Button from "@/components/input/Button.vue";

</script>

<template>
    <audio/>
    <div class="musicPlayerGrid">
        <div class="album_cover">
            <img src="/img/Untitled.png" alt=""></img>
        </div>
        <div class="controls">
            <div class="sliders">
                <div class="timeline"/>
                <div class="volume"/>
            </div>
            <div class="buttons">
                <div class="rewind"/>
                <div class="playPause"/>
                <div class="fastforward"/>
            </div>
        </div>
    </div>
</template>

<style scoped>
.musicPlayerGrid {
    display: grid;
    grid-gap: 5px;
    grid-template-rows: repeat(4, 1fr);
    background-color: blue;
    align-items: stretch; /* rows (block axis) */
    justify-items: stretch; /* columns (inline axis) */
    padding: 5px;
}
img {
    width: 100%;
}

.album_cover {
    grid-row: 1 / span 3;
    background-color: grey;
    box-sizing: border-box;
}

.controls {
    width: 100%;
    grid-row: 4 / span 1;
    box-sizing: border-box;

    display: grid;
    grid-template-rows: repeat(4, 1fr);
    grid-gap: 5px;
}

.sliders {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    grid-gap: 5px;
}

.timeline {
    grid-column: 1;
    background-color: white;
}

.volume {
    grid-column: 2;
    background-color: white;
}

.buttons {
    background-color: black;
    grid-row: 2 / -1;

    display: grid;
    grid-template-columns: repeat(4, 1fr);
    grid-gap: 5px;
}

.rewind {
    grid-column: 1;
    background-color: grey;
}
.fastforward {
    grid-column: 4;
    background-color: grey;
}
.playPause {
    grid-column: 2/span 2;
    background-color: grey;
}
</style>
