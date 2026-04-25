<script setup>
import Project from "./Project.vue";
</script>

<template>
    <main>
        <div class="no-print w-full h-20"></div>
        <div class="a4page justify-between">
            <section>
                <div class="flex flex-col sm:flex-row sm:justify-between">
                    <h1 class="flex-1 name">Adam French</h1>
                    <div class="contact-details">
                        <p>London, United Kingdom</p>
                        <p>+447563266931</p>
                        <p>adam.a.french@outlook.com</p>
                        <p>
                            <a href="https://www.adam-french.co.uk">
                                www.adam-french.co.uk
                            </a>
                        </p>
                    </div>
                </div>
            </section>

            <section>
                <h2>Profile</h2>
                <p contenteditable="true">
                    Full Stack Developer with a First Class Honours degree in
                    Computer Science with Mathematics from the University of
                    Leeds (81.1%) and a year abroad at the University of
                    Waterloo. Proficient in full-stack development, systems
                    programming, and CI/CD automation. Strong problem-solving
                    and organisational skills. Eager to contribute to a
                    collaborative engineering team, apply strong academic
                    foundations to real-world problems, and grow through
                    hands-on experience.
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
                        <strong>Frontend / Web Design</strong><br /><small
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
                            Nginx, Vue, Postgres, Docker, Go, Python, Rust,
                            Wasm, Git Actions, JWT Auth
                        </small>
                        <small>2025</small>
                    </template>
                    <p contenteditable="true">
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
                    <p contenteditable="true">
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
                    <p contenteditable="true">
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
                    <p contenteditable="true">
                        Research project on Mobile Automata with data
                        visualisation and academic presentation. Delivered
                        within a tight deadline in collaboration with academic
                        mentors.
                    </p>
                </Project>
            </section>

            <div class="w-full flex flex-col sm:flex-row gap-5">
                <section class="flex-1">
                    <h2>
                        <a
                            href="https://www.adam-french.co.uk/pdf/transcript.pdf"
                        >
                            University of Leeds
                        </a>
                    </h2>
                    <div
                        class="flex-row flex place-content-between m-auto place-items-center"
                    >
                        <small>First Class Honours (81.1%)</small>
                        <small>Sep 2021 – Jun 2025</small>
                    </div>
                    <small>BSc Computer Science with Mathematics </small>
                    <ul class="list-disc list-inside">
                        <li>Algorithms & Data Structures I & II</li>
                        <li>Compiler Design and Construction</li>
                        <li>Formal Languages & Finite Automata</li>
                        <li>Graph Algorithms & Complexity Theory</li>
                        <li>
                            Machine Learning, Databases, Computer Processors
                        </li>
                    </ul>
                </section>
                <section class="flex-1">
                    <h2>University of Waterloo</h2>
                    <div
                        class="flex-row flex place-content-between m-auto place-items-center"
                    >
                        <small>Year abroad</small>
                        <small>Sep 2023 – Apr 2024</small>
                    </div>
                    <ul class="list-disc list-inside">
                        <li>Applied Cryptography</li>
                        <li>Introduction to Computer Graphics</li>
                        <li>
                            Introduction to Rings and Fields with Applications
                        </li>
                    </ul>
                </section>
            </div>
        </div>

        <div class="no-print w-full h-20"></div>

        <div class="a4page gap-10">
            <section>
                <h2>Experience</h2>
                <Project>
                    <template #left>
                        <h4>Hospitality</h4>
                    </template>
                    <template #top>
                        <small>Cashier, Bartender, Waiter</small>
                        <small>Jan 2018 – Dec 2023</small>
                    </template>
                    <p contenteditable="true">
                        Worked at <strong>Belgrave Music Hall</strong>,
                        <strong>The Crown and Anchor</strong>, and
                        <strong>BFI Riverfront Kitchen</strong>. Developed
                        communication, composure under pressure, and reliability
                        in customer-facing roles.
                    </p>
                </Project>
            </section>
            <div class="w-full flex flex-col sm:flex-row gap-5">
                <section class="flex-1">
                    <h2>Soft Skills</h2>
                    <ul class="list-disc list-inside">
                        <li>Communication & collaboration</li>
                        <li>Attention to detail</li>
                        <li>Problem solving</li>
                        <li>Adaptability</li>
                        <li>Time management</li>
                    </ul>
                </section>
                <section class="flex-1">
                    <h2>Interests</h2>
                    <ul class="list-disc list-inside">
                        <li>Leetcode</li>
                        <li>Learning Mandarin</li>
                        <li>Rhythm Games</li>
                        <li>Climbing, Gym</li>
                        <li>Board games, Meetup.com</li>
                    </ul>
                </section>
            </div>
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
    --font-size-text: 1em;
    --font-size-small: 0.9em;
    --font-size-heading: 1.5em;
    --font-size-subheading: 1.5em;
    --font-size-subsubheading: 1.3em;
}

/* A4 Page */
.a4page {
    line-height: 1.6;
    font-family: var(--font-text);
    width: 210mm;
    height: 297mm;
    padding: 10mm;
    box-sizing: border-box;
    background-color: var(--background);
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
    border: 1px solid var(--primary);
    overflow: hidden;
    margin: auto auto;
    display: flex;
    flex-direction: column;
}

.name {
    font-size: 2em;
    font-weight: bold;
    justify-content: center;
    align-content: center;
}

/* Component Styling */
main {
    padding: 0px;
    display: flex;
    flex-direction: column;
    height: fit-content;
    background-color: white;
}

strong {
    font-weight: 900;
}

span {
    height: 2em;
}

h1,
h2,
h3,
h4 {
    margin: 0px 0px 0.2em 0px;
    border: none;
    color: var(--primary);
    font-family: var(--font-heading);
    text-transform: uppercase;
}

h1 {
    font-size: var(--font-size-heading);
}

h2 {
    border-bottom: 1px solid var(--primary);
    font-size: var(--font-size-subheading);
}

h3,
h4 {
    font-size: var(--font-size-subsubheading);
}

a:hover {
    color: var(--tertiary);
}

a {
    background-color: transparent;
    color: var(--secondary);
    font-family: inherit;
    font-size: var(--font-size-text);
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

p[contenteditable] {
    outline: none;
    cursor: text;
}

p[contenteditable]:focus {
    outline: none;
    background-color: rgba(0, 0, 255, 0.05);
}

@media (max-width: 640px) {
    .a4page {
        width: 100%;
        height: auto;
        overflow: visible;
        box-shadow: none;
        border: none;
    }

    .skills-grid {
        grid-template-columns: 1fr;
    }
}
</style>
