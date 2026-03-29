<script setup>
import Project from "./Project.vue";
</script>

<template>
    <main>
        <div class="no-print w-full h-20"></div>
        <div class="a4page justify-between">
            <section>
                <div class="flex flex-row justify-between">
                    <h1 class="name">Adam French</h1>
                    <div class="contact-details text-right">
                        <p>+447563266931</p>
                        <p>adam.a.french@outlook.com</p>
                        <h4>
                            <a href="https://www.adam-french.co.uk">
                                www.adam-french.co.uk
                            </a>
                        </h4>
                    </div>
                </div>
            </section>

            <section>
                <h2>Profile</h2>
                <p>
                    First Class Honours graduate in Computer Science with
                    Mathematics from the University of Leeds (81.1%), with a
                    year abroad at the University of Waterloo. Proficient in
                    full-stack development, systems programming, and CI/CD
                    automation. Eager to contribute to a collaborative
                    engineering team, apply strong academic foundations to
                    real-world problems, and grow through hands-on experience.
                </p>
            </section>

            <section>
                <h2>Skills</h2>
                <div class="skills-grid">
                    <div>
                        <strong>Languages</strong><br /><small
                            >Go, Rust, Python, JavaScript / TypeScript,
                            SQL</small
                        >
                    </div>
                    <div>
                        <strong>Frontend</strong><br /><small
                            >Vue, React / Redux, Svelte, Tailwind CSS,
                            WebAssembly</small
                        >
                    </div>
                    <div>
                        <strong>Backend / Infra</strong><br /><small
                            >Nginx, Docker, PostgreSQL, SQLite, JWT Auth, Git
                            Actions</small
                        >
                    </div>
                </div>
            </section>

            <section>
                <h2>Projects</h2>

                <Project class="border-b border-dotted">
                    <template v-slot:left>
                        <h4>
                            <a
                                href="https://www.adam-french.co.uk/gitea/adamf/web_server.git"
                            >
                                web_server.git
                            </a>
                        </h4>
                    </template>
                    <template v-slot:top>
                        <small>
                            Nginx, Vue, Postgres, Docker, Go, Python, Rust →
                            Wasm, Git Actions, JWT Auth
                        </small>
                        <small>2025</small>
                    </template>
                    <p>
                        Self-hosted personal website with a fully automated
                        CI/CD pipeline. Iterated across diverse tech stacks
                        including Svelte, React/Redux, SQLite, Rust Actix, and
                        Deno.
                    </p>
                </Project>
                <Project class="border-b border-dotted">
                    <template v-slot:left>
                        <h4>
                            <a
                                href="https://www.adam-french.co.uk/gitea/adamf/tour.git"
                            >
                                tour.git
                            </a>
                        </h4>
                    </template>
                    <template v-slot:top>
                        <small>Rust</small>
                        <small>2026</small>
                    </template>
                    <p>
                        CLI tool for building and navigating interactive code
                        tutorials, with version-traversal semantics inspired by
                        Git.
                    </p>
                </Project>
                <Project class="border-b border-dotted">
                    <template v-slot:left>
                        <h4>
                            <a
                                href="https://www.adam-french.co.uk/gitea/adamf/rust-raytracer.git"
                            >
                                rust-raytracer.git
                            </a>
                        </h4>
                    </template>
                    <template v-slot:top>
                        <small>Rust, Linear Algebra, Multithreading</small>
                        <small>2023</small>
                    </template>
                    <p>
                        Parallelised recursive ray tracer for realistic 3D
                        rendering. Emphasised algorithmic efficiency and
                        low-level memory management in Rust.
                    </p>
                </Project>
                <Project>
                    <template #left>
                        <h4>
                            <a
                                class="text-center w-full"
                                href="https://community.wolfram.com/groups/-/m/t/3210947"
                            >
                                Wolfram Summer School
                            </a>
                        </h4>
                    </template>
                    <template #top>
                        <small>Wolfram Mathematica</small>
                        <small>2024</small>
                    </template>
                    <p>
                        Research project on Mobile Automata with data
                        visualisation and academic presentation. Delivered
                        within a tight deadline in collaboration with academic
                        mentors.
                    </p>
                </Project>
            </section>

            <section>
                <h2>Education</h2>
                <div class="w-full h-fit flex-row flex gap-5">
                    <div class="flex-1 border-r border-dotted pr-3">
                        <h3>
                            <a
                                href="https://www.adam-french.co.uk/pdf/transcript.pdf"
                            >
                                University of Leeds
                            </a>
                        </h3>
                        <div
                            class="flex-row flex place-content-between m-auto place-items-center"
                        >
                            <small>First Class Honours (81.1%)</small>
                            <small>2021–2025</small>
                        </div>
                        <small>BSc Computer Science with Mathematics </small>
                        <ul class="list-disc list-inside">
                            <li>Algorithms & Data Structures I & II</li>
                            <li>Compiler Design and Construction</li>
                            <li>Formal Languages & Finite Automata</li>
                            <li>Graph Algorithms & Complexity Theory</li>
                            <li>
                                Machine Learning · Databases · Computer
                                Processors
                            </li>
                        </ul>
                    </div>
                    <div class="flex-1 pl-3">
                        <h3>University of Waterloo</h3>
                        <div
                            class="flex-row flex place-content-between m-auto place-items-center"
                        >
                            <small>Year abroad</small>
                            <small>2023–2024</small>
                        </div>
                        <ul class="list-disc list-inside">
                            <li>Applied Cryptography</li>
                            <li>Introduction to Computer Graphics</li>
                            <li>
                                Introduction to Rings and Fields with
                                Applications
                            </li>
                        </ul>
                    </div>
                </div>
            </section>
        </div>

        <div class="no-print w-full h-20"></div>

        <div class="a4page gap-10">
            <section>
                <h2>Experience</h2>
                <Project>
                    <template #left>
                        <p>Hospitality</p>
                    </template>
                    <template #top>
                        <small>Cashier, Bartender, Waiter</small>
                        <small>2018–2023</small>
                    </template>
                    <p>
                        Worked at <em>Belgrave Music Hall</em>,
                        <em>The Crown and Anchor</em>, and
                        <em>BFI Riverfront Kitchen</em>. Developed
                        communication, composure under pressure, and reliability
                        in customer-facing roles.
                    </p>
                </Project>
            </section>
            <section>
                <h2>Interests</h2>
                <ul class="list-disc list-inside">
                    <li>Leetcode — daily competitive problem solving</li>
                    <li>Learning Mandarin</li>
                    <li>Rhythm Games</li>
                    <li>Climbing · Gym</li>
                    <li>Board games · Meetup.com</li>
                </ul>
            </section>
        </div>

        <div class="no-print w-full h-20"></div>
    </main>
</template>

<style scoped>
/* Fonts */
@font-face {
    font-family: "big_noodle_titling";
    src: url("/fonts/big_noodle_titling.woff2") format("woff2");
    font-weight: normal;
    font-style: normal;
    font-display: swap;
}

@font-face {
    font-family: "CreatoDisplay";
    src: url("/fonts/CreatoDisplay-Bold.woff2") format("woff2");
    font-weight: normal;
    font-style: normal;
    font-display: swap;
}

/* Variables */
* {
    --primary: black;
    --secondary: #0000ff;
    --tertiary: #ff0000;
    --quaternary: #cccccc;
    --background: white;

    --font-heading: big_noodle_titling;
    --font-text: CreatoDisplay;
    --font-size-name: 2.5em;
    --font-size-text: 100%;
    --font-size-small: 0.9em;
    --font-size-heading: 2.1em;
    --font-size-subheading: 1.7em;
    --font-size-subsubheading: 1.4em;
}

/* A4 Page */
.a4page {
    line-height: 1.6;
    font-family: var(--font-text);
    width: 210mm;
    height: 297mm;
    padding: 5mm;
    box-sizing: border-box;
    background-color: var(--background);
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
    border: 1px solid var(--primary);
    overflow: hidden;
    margin: auto auto;
    display: flex;
    flex-direction: column;
}

/* Component Styling */
main {
    padding: 0px;
    display: flex;
    flex-direction: column;
    height: fit-content;
    background-color: white;
}

span {
    height: 2em;
}

h1,
h2,
h3,
h4 {
    margin: 0px;
    border: none;
    color: var(--primary);
    font-family: var(--font-heading);
    text-transform: capitalize;
}

h1 {
    font-size: var(--font-size-heading);
}

h2 {
    border-bottom: 1px solid var(--primary);
    font-size: var(--font-size-subheading);
}

h3 {
    font-size: var(--font-size-subsubheading);
}

a:hover {
    color: var(--tertiary);
}
a {
    background-color: transparent;
    color: var(--secondary);
}

p {
    margin-bottom: 0.2em;
    color: var(--primary);
    font-size: var(--font-size-text);
}

table {
    color: var(--secondary);
    border-collapse: collapse;
    border: 1px solid black;
}

td {
    color: var(--secondary);
    border-top: 1px solid var(--tertiary);
    padding: 1px 10px 1px 10px;
    font-size: var(--font-size-text);
    text-align: left;
}

th {
    color: var(--secondary);
    border: 2px solid var(--tertiary);
    padding: 1px 0px 1px 7px;
    font-family: var(--font-heading);
    font-size: var(--font-size-subsubheading);
    background-color: var(--quaternary);
    text-align: left;
}

@media print {
    .no-print {
        display: none !important;
    }
}

small {
    font-size: var(--font-size-small);
    color: var(--primary);
}

ul {
    font-size: var(--font-size-small);
    margin: 0;
    padding-left: 1.2em;
}

li {
    font-size: var(--font-size-small);
    color: var(--primary);
}

.skills-grid {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 0.3em 1em;
    margin-bottom: 0.2em;
}
</style>
