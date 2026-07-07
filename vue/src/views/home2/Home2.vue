<template>
    <div class="menu">
        <div class="header">
            <div class="sort-select"></div>
            <div class="mix-select"></div>
            <div class="levelbar"></div>
        </div>
        <div class="footer">
            <div class="footer-difficulty"></div>
            <div class="footer-controls">
                <div class="control-container">
                    <p class="control-button">F2</p>
                    <p class="control-text">Help</p>
                </div>
                <div class="control-container">
                    <p class="control-button">T</p>
                    <p class="control-text">Sort</p>
                </div>
                <div class="control-container">
                    <p class="control-button">F1</p>
                    <p class="control-text">-/+ Favorite</p>
                </div>
                <div class="control-container">
                    <p class="control-button">R</p>
                    <p class="control-text">Customize</p>
                </div>
                <div class="control-container">
                    <p class="control-button">Enter</p>
                    <p class="control-text">Ok</p>
                </div>
                <div class="control-container">
                    <p class="control-button">Esc</p>
                    <p class="control-text">Back</p>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
div {
    opacity: 80%;
}

.menu {
    position: absolute;
    inset: 0;
    width: 1920px;
    height: 1080px;
    background-image: url("/img/ss_c523b1c1311d195e10b1a96c70db1b85c20a8bda.1920x1080.jpg");
    background-color: rgba(255, 255, 255, 0.5);
    background-blend-mode: lighten;
    background-position: center;
    background-repeat: no-repeat;
}

.header {
    display: grid;
    grid-template-areas: "sort sort sort sort sort sort sort sort sort mix mix levelbar levelbar levelbar levelbar";
    flex-direction: row;
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 10%;
    background-color: rgba(128, 128, 128, 0.5);
}

.levelbar {
    grid-area: levelbar;
    background-color: green;
    flex: 1;
}

.mix-select {
    grid-area: mix;
    flex: 1;
    background-color: red;
}

.sort-select {
    grid-area: sort;
    flex: 1;
    background-color: blue;
}

.footer {
    position: absolute;
    bottom: 0;
    left: 0;
    width: 100%;
    height: 15%;
    display: flex;
    flex-direction: column;
}

.footer-difficulty {
    height: 58%;
    background-color: red;
}

.footer-controls {
    background-color: blue;
    height: 42%;
    display: flex;
    flex-direction: row;
    justify-content: flex-end;
    align-items: center;
    gap: 0.75rem;
    padding: 0 0.75rem;
}

.control-container {
    display: flex;
    align-items: center;
    gap: 0.35rem;
}

.control-text {
    font-family: sans-serif;
    font-style: italic;
}

.control-button {
    background-color: gray;
    border-radius: 0.35rem;
    padding: 0.15rem 0.5rem;
}
</style>
