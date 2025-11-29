<template>
    <nav class="no-print flex-row">
        <RouterLink to="/"><h1>HOME</h1></RouterLink>
        <RouterLink to="/cv"><h1>CV</h1></RouterLink>
    </nav>
</template>

<style scoped>
nav {
    height: 10vh;
}
a {
    rotate: -50deg;
    height: 100%;
    background: white;
    padding: 4px;
    border: 2px solid black;
}
</style>
