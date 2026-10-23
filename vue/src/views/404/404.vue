<!--
  Catch-all route for unmatched paths under `/`.
-->
<template>
  <main class="flex flex-col items-center">
    <div
      class="page-a4 items-center bdr-1 flex flex-col relative overflow-auto"
    >
      <h1>404</h1>
      <RouterLink to="/" class="bdr-2" aria-label="Back to the home page">
        <img src="/img/memes/epic.jpeg" alt="" class="block max-w-full" />
      </RouterLink>
      <h1>Click her, she will take you home</h1>
    </div>
  </main>
</template>
