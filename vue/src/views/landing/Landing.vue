<script setup>
/**
 * Route `/` — the professional front door. Deliberately plain: a short bio and a
 * row of links out to the CV, the personal home page (/stp), the projects that
 * run as their own containers, and social profiles.
 * Statically imported by the router (it is the first paint).
 */
import Link from "@/components/text/Link.vue";
import InlineLink from "@/components/text/InlineLink.vue";
import Header from "@/components/text/Header.vue";
import Paragraph from "@/components/text/Paragraph.vue";

// The projects are nginx routes to other containers, not SPA routes, so they
// must be plain hrefs: a RouterLink would resolve them client-side to the 404.
const projects = [
  { name: "Jobs", href: "/jobs/" },
  { name: "Tracker", href: "/tracker/" },
  { name: "USACO", href: "/usaco/" },
];

const links = [
  { name: "GitHub", href: "https://github.com/SteveThePug" },
  { name: "Gitea", href: "/gitea/explore/repos" },
  { name: "Spotify", href: "https://open.spotify.com/user/stevethepug" },
];
</script>

<template>
  <main class="flex justify-center px-4 py-16">
    <div class="max-w-xl w-full flex flex-col gap-12">
      <section>
        <Header>Adam French</Header>
        <Paragraph>
          Junior software engineer focused on full-stack development, systems
          programming, and infrastructure. First Class Honours in Computer
          Science with Mathematics from Leeds and Waterloo.
        </Paragraph>
      </section>

      <section>
        <Header>About</Header>
        <Paragraph>
          This website is self-hosted and has a lot more on it than it needs to.
          Please have a look at my
          <InlineLink to="/cv">CV</InlineLink> for a full breakdown of my
          experience, projects, and skills. Please visit
          <InlineLink to="/stp">STP</InlineLink> for the prefered but less
          professional experience.
        </Paragraph>
      </section>

      <nav class="navRow flex flex-row flex-wrap gap-4 justify-around">
        <Link to="/cv"> CV </Link>
        <Link to="/stp"> STP </Link>
        <Link v-for="p in projects" :key="p.name" :href="p.href">
          {{ p.name }}
        </Link>
        <Link href="mailto:adam.a.french@outlook.com"> Email </Link>
        <Link v-for="link in links" :key="link.name" :href="link.href">
          {{ link.name }}
        </Link>
      </nav>
    </div>
  </main>
</template>

<style scoped>
.navRow > a {
  padding: 0.5rem 1rem;
  border: 1px solid currentColor;
}
</style>
