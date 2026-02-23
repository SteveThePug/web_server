<script setup>
import Project from "./Project.vue";
</script>

<template>
    <main>
        <div class="a4page">
            <div class="flex flex-row justify-between">
                <h1 class="name">Adam French</h1>
                <!-- <a href="covers.html"><img width=25 height=50 src="img/rune.png"></a> -->
                <div class="contact-details text-right">
                    <p>+447563266931</p>
                    <p>adam.a.french@outlook.com</p>
                    <a href="https://www.adam-french.co.uk">
                        www.adam-french.co.uk
                    </a>
                </div>
            </div>

            <h2>Profile</h2>
            <p>
                Recent Computer Science with Mathematics (International)
                graduate from the University of Leeds, awarded First Class
                Honours (81.1%). Strong foundation in full-stack software
                development, CI/CD workflows, and modern programming languages.
                Experienced in creating scalable, maintainable systems and
                motivated by solving complex technical problems. Enthusiastic
                about working within organisations that promote innovation,
                collaboration, and positive social impact.
            </p>

            <h2>Projects</h2>

            <Project class="border-b border-dotted">
                <template v-slot:left>
                    <a
                        href="https://www.adam-french.co.uk/gitea/adamf/web_server.git"
                    >
                        web_server.git
                    </a>
                </template>
                <template v-slot:top>
                    <small>
                        Nginx, Vue, Postgres, Docker, Go, Python, Rust -> Wasm,
                        Git Actions, JWT Auth
                    </small>
                    <small>2025</small>
                </template>
                <p>
                    Developed and self-hosted a personal website with a fully
                    automated maintenance CI/CD pipeline. Experimented with
                    diverse tech stacks including Svelte, React/Redux, SQLite,
                    Rust Actix, and Deno.
                </p>
            </Project>
            <Project class="border-b border-dotted">
                <template v-slot:left>
                    <a
                        href="https://www.adam-french.co.uk/gitea/adamf/tour.git"
                    >
                        tour.git
                    </a>
                </template>
                <template v-slot:top>
                    <small>Rust</small>
                    <small>2026</small>
                </template>
                <p>
                    Created a command-line tool for building and viewing
                    interactive code tutorials. Designed functionality analogous
                    to Git for intuitive version traversal and educational use.
                </p>
            </Project>
            <Project class="border-b border-dotted">
                <template v-slot:left>
                    <a
                        href="https://www.adam-french.co.uk/gitea/adamf/rust-raytracer.git"
                    >
                        rust-raytracer.git
                    </a>
                </template>
                <template v-slot:top>
                    <small>Rust, Linear Algebra, Multithreading</small>
                    <small>2023</small>
                </template>
                <p>
                    Built a parallelised, recursive ray tracer for realistic 3D
                    rendering as part of a university module. Focused on
                    algorithmic efficiency and low-level memory management in
                    Rust.
                </p>
            </Project>
            <Project>
                <template #left>
                    <p>
                        <a
                            class="text-center w-full"
                            href="https://community.wolfram.com/groups/-/m/t/3210947"
                        >
                            Wolfram Summer School
                        </a>
                    </p>
                </template>
                <template #top>
                    <small>Wolfram Mathematica</small>
                    <small>2024</small>
                </template>
                <p>
                    Designed and implemented a research project on Mobile
                    Automata, including data visualisation and presentation of
                    findings. Completed the project within a short deadline and
                    collaborated with academic mentors to refine outcomes.
                </p>
            </Project>
            <h2>University & Modules</h2>
            <div class="w-full h-fit flex-row flex gap-5">
                <div class="flex-1 border-r border-dotted pr-3">
                    <h3>University of Leeds</h3>
                    <div
                        class="flex-row flex place-content-between m-auto place-items-center"
                    >
                        <small> 81.1% (First Class Honours)</small>
                        <small> 2021-2025 </small>
                    </div>
                    <small>BSc Computer Science with Mathematics </small>
                    <ul>
                        <li>Procedural & Object Oriented Programming,</li>
                        <li></li>
                        <li>Algorithms and Data Structures I & II</li>

                        <li>Databases</li>
                        <li>Computer Processors</li>
                        <li>Compiler Design and Construction</li>

                        <li>Formal Languages and Finite Automata</li>
                        <li>Probability and Statistics I</li>
                        <li>Machine Learning</li>
                        <li>Graph Algorithms & Complexity Theory</li>
                    </ul>
                </div>
                <div class="flex-1 pl-3">
                    <h3>The University of Waterloo</h3>
                    <div
                        class="flex-row flex place-content-between m-auto place-items-center"
                    >
                        <small>---</small>
                        <small> 2023-2024 </small>
                    </div>
                    <div class="flex-row flex place-content-between"></div>
                    <ul>
                        <li>Applied Cryptography</li>
                        <li>Introduction to Computer Graphics</li>
                        <li>
                            Introduction to Rings and Fields with Applications
                        </li>
                    </ul>
                </div>
            </div>
        </div>
        <!-- <div class="a4page"> -->
        <!--     <h2>Experience</h2> -->
        <!--     <Project> -->
        <!--         <template #left> -->
        <!--             <p>Hospitality</p> -->
        <!--         </template> -->
        <!--         <template #top> -->
        <!--             <small>Cashier, Bartender, Waiter</small> -->
        <!--             <small>2018-2023</small> -->
        <!--         </template> -->
        <!--         <p> -->
        <!--             Worked at venues including: -->
        <!--             <em>Belgrave Music Hall</em>, -->
        <!--             <em>The Crown and Anchor Eastbourne</em>, -->
        <!--             <em>To The Rise Bakery</em>, -->
        <!--             <em>BFI Riverfront Kitchen</em>. -->
        <!--         </p> -->
        <!--     </Project> -->
        <!--     <h2>Commitments</h2> -->
        <!--     <ul> -->
        <!--         <li>Gym</li> -->
        <!--         <li>Climbing</li> -->
        <!--         <li>Meetup.com</li> -->
        <!--         <li>Boardgames</li> -->
        <!--         <li>Leetcode</li> -->
        <!--         <li>Learning Mandarin</li> -->
        <!--     </ul> -->
        <!-- </div> -->
    </main>
</template>

<style scoped>
/* Fonts */
@font-face {
    font-family: "big_noodle_titling";
    src: url("/fonts/big_noodle_titling.ttf") format("truetype");
    font-weight: normal;
    font-style: normal;
}

@font-face {
    font-family: "CreatoDisplay";
    src: url("/fonts/CreatoDisplay-Bold.otf") format("opentype");
    font-weight: normal;
    font-style: normal;
}

/* Variables */
* {
    /* Black - White */
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
    /* Standard A4 width */
    height: 297mm;
    /* Standard A4 height */
    padding: 5mm;
    box-sizing: border-box;
    background-color: var(--background);
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
    border: 1px solid var(--primary);
    overflow: hidden;
    /* Enables scrolling when content exceeds height */
    margin: auto auto;
    /* Centers the page horizontally */
}

/* Component Styling */
main {
    padding: 0px;
    display: flex;
    flex-direction: column;
    height: fit-content;
}

span {
    height: 2em;
}

h1,
h2,
h3,
h4 {
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
    /* border: 2px solid var(--tertiary); */
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
}

li {
    font-size: var(--font-size-small);
    color: var(--primary);
}
</style>
