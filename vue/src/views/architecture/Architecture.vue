<script setup>
/**
 * Route `/architecture` — how the whole site fits together, drawn.
 *
 * One long page of hand-written SVG diagrams, grouped into five sections
 * (sections/*Section.vue), each of which lists its diagrams
 * (diagrams/<section>/*.vue). Every diagram sits in a <DiagramFigure> and draws
 * with the `dg-` classes in diagram.css; DiagramDefs supplies the shared
 * arrowheads. No chart or graph library: the diagrams are static pictures of
 * the code, so plain SVG is the smallest thing that does the job.
 *
 * The diagrams describe the code as it was when they were drawn. When a
 * behaviour they show changes, redraw the diagram in the same commit.
 */
import "./diagram.css";
import Header from "@/components/text/Header.vue";
import Paragraph from "@/components/text/Paragraph.vue";
import DiagramDefs from "./DiagramDefs.vue";
import InfraSection from "./sections/InfraSection.vue";
import EdgeSection from "./sections/EdgeSection.vue";
import BackendSection from "./sections/BackendSection.vue";
import FrontendSection from "./sections/FrontendSection.vue";
import ServicesSection from "./sections/ServicesSection.vue";

const SECTIONS = [
    { id: "infra", label: "Infrastructure & deployment" },
    { id: "edge", label: "The edge: nginx" },
    { id: "backend", label: "Go backend" },
    { id: "frontend", label: "Vue frontend" },
    { id: "services", label: "Python & the other services" },
];
</script>

<template>
    <main class="architecture flex justify-center px-4 py-10">
        <DiagramDefs />
        <div class="max-w-6xl w-full flex flex-col gap-10">
            <section>
                <Header>How this site works</Header>
                <Paragraph>
                    Everything behind this website, drawn. It is a couple of
                    dozen Docker containers on one Ubuntu VM: nginx in front, a
                    Go GraphQL backend, a Vue single-page app, a Python API that
                    holds the Claude call and the scrapers, and a ring of
                    self-hosted services (Gitea, SilverBullet notes, Garage
                    object storage, Icecast radio, a job tracker) around them.
                    Each diagram can be expanded to full screen.
                </Paragraph>
                <nav class="toc">
                    <a v-for="s in SECTIONS" :key="s.id" :href="`#${s.id}`">
                        {{ s.label }}
                    </a>
                </nav>
            </section>

            <InfraSection />
            <EdgeSection />
            <BackendSection />
            <FrontendSection />
            <ServicesSection />
        </div>
    </main>
</template>

<style scoped>
.toc {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    padding: 0.5rem 0.25rem;
}
.toc a {
    padding: 0.2rem 0.75rem;
    border: 1px solid var(--color-quaternary);
    text-decoration: none;
}
.toc a:hover {
    border-color: var(--color-primary);
}
.architecture :deep(section) {
    scroll-margin-top: 4rem;
}
</style>
