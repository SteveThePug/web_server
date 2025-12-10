<template>
    <div>
        <RouterLink to="/cv">
            <h1>CV</h1>
        </RouterLink>
        <RouterLink to="/admin">
            <h1>ADMIN</h1>
        </RouterLink>
    </div>
</template>

<style scoped>
div {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    padding: 2px;
}
p {
    background: var(--bg_primary);
}
</style>
