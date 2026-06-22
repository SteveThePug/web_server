<script setup>
import Project from "./Project.vue";
</script>

<template>
    <main class="cv-template">
        <div class="no-print w-full h-20"></div>
        <div class="a4page justify-between">
            <section>
                <div class="flex flex-col sm:flex-row sm:justify-between">
                    <h1 class="name">Adam French</h1>
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
            <section class="w-full flex flex-col sm:flex-row gap-5">
                <h2>Profile</h2>
                <div class="flex-1">
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
                </div>
                <div class="flex-1">
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
        </div>

        <div class="no-print w-full h-20"></div>

        <div class="a4page gap-10">
            <section>
                <h2>University Academy of Engineering Southbank</h2>
                <div
                    class="flex-row flex place-content-between m-auto place-items-center"
                >
                    <small>Secondary & Sixth Form Education</small>
                    <small>Sep 2014 – Jun 2021</small>
                </div>
                <div class="w-full flex flex-col sm:flex-row gap-5">
                    <div class="flex-1">
                        <strong>A-Levels</strong>
                        <div
                            style="
                                display: grid;
                                grid-template-columns: 1fr auto;
                                column-gap: 1em;
                            "
                        >
                            <span>Mathematics</span><span>A*</span>
                            <span>Further Mathematics</span><span>A*</span>
                            <span>Physics</span><span>A*</span>
                        </div>
                    </div>
                    <div class="flex-1">
                        <strong>GCSEs</strong>
                        <div
                            style="
                                display: grid;
                                grid-template-columns: 1fr auto;
                                column-gap: 1em;
                            "
                        >
                            <span>Mathematics</span><span>9</span>
                            <span>English Literature</span><span>9</span>
                            <span>Physics</span><span>9</span>
                            <span>English Language</span><span>8</span>
                            <span>Biology</span><span>8</span>
                            <span>Chemistry</span><span>8</span>
                            <span>Computer Science</span><span>7</span>
                        </div>
                    </div>
                </div>
            </section>

            <section>
                <h2>Experience</h2>

                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>Daisy Green Holland Park</h4>
                    </template>
                    <template #top>
                        <small>Barista & Front of House</small>
                        <small>May 2026 – Present</small>
                    </template>
                    <p contenteditable="true">
                        Working at an all-day café in Holland Park, preparing
                        coffee and serving food in a fast-paced environment.
                    </p>
                </Project>
                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>Belgrave Music Hall</h4>
                    </template>
                    <template #top>
                        <small>Bartender & Waiter</small>
                        <small>2021–2025</small>
                    </template>
                    <p contenteditable="true">
                        Served food and drinks in a high-volume live-music venue
                        in Leeds. Handled busy weekend shifts, managed multiple
                        tables simultaneously, and maintained a calm, friendly
                        demeanour during peak hours.
                    </p>
                </Project>
                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>The Crown and Anchor</h4>
                    </template>
                    <template #top>
                        <small>Bartender & Waiter</small>
                        <small>2020–2021</small>
                    </template>
                    <p contenteditable="true">
                        Worked front-of-house at a busy pub, pulling pints,
                        taking orders, and ensuring a welcoming atmosphere.
                        Built rapport with regulars and adapted quickly to
                        changing priorities during service.
                    </p>
                </Project>
                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>To The Rise Bakery</h4>
                    </template>
                    <template #top>
                        <small>Barista & Front of House</small>
                        <small>2020–2021</small>
                    </template>
                    <p contenteditable="true">
                        Worked at a bakery in Eastbourne, preparing freshly
                        baked goods for display, operating the coffee machine,
                        and keeping bakery equipment spotless. Built rapport
                        with regulars and provided attentive, friendly service.
                    </p>
                </Project>
                <Project>
                    <template #left>
                        <h4>BFI Riverfront Kitchen</h4>
                    </template>
                    <template #top>
                        <small>Cashier & Waiter</small>
                        <small>2018–2020</small>
                    </template>
                    <p contenteditable="true">
                        Operated the till, served customers, and helped
                        coordinate table service at a café on London's South
                        Bank. Developed strong cash-handling accuracy and
                        customer interaction skills in a fast-paced environment.
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
