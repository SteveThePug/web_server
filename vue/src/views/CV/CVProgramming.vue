<script setup>
/**
 * The live CV template — the only one currently listed in CV.vue.
 *
 * Both sheets are fixed A4 with overflow hidden (cv-shared.css), so anything
 * added here has to be checked in print preview: text that doesn't fit is
 * silently clipped, not pushed onto another page.
 */
import Project from "./Project.vue";
</script>

<template>
    <main class="cv-template">
        <div class="no-print w-full h-20"></div>
        <div class="cv-page justify-around">
            <section>
                <div class="flex flex-col sm:flex-row sm:justify-between">
                    <h1 class="name">Adam French</h1>
                    <div class="contact-details">
                        <p>London, United Kingdom</p>
                        <p>+44 7563 266931</p>
                        <p>adam.a.french@outlook.com</p>
                        <p>
                            <a href="https://www.adam-french.co.uk">
                                www.adam-french.co.uk
                            </a>
                        </p>
                        <p>
                            <a href="https://github.com/SteveThePug">
                                github.com/SteveThePug
                            </a>
                        </p>
                    </div>
                </div>
            </section>

            <section>
                <h2>Profile</h2>
                <p>
                    Software engineer with a First Class BSc in Computer
                    Science with Mathematics. I build and run full-stack systems
                    end to end, from Go services and Vue frontends to the
                    Docker, Nginx and CI/CD infrastructure beneath them: around
                    twenty services in production at
                    <a href="https://www.adam-french.co.uk/architecture">
                        adam-french.co.uk</a
                    >. My <a href="/pdf/dissertation.pdf">dissertation</a>
                    proved <a href="/automata">Mobile Automata</a> are
                    computationally equivalent to Turing Machines. I use AI in
                    my workflow (Claude Code) and in what I ship (the Claude
                    API).
                </p>
            </section>

            <section>
                <h2>Technical Skills</h2>
                <p>
                    <strong>Languages:</strong>
                    Go, Python, Rust, C++, JavaScript / TypeScript, SQL, Bash
                </p>
                <p>
                    <strong>Backend & Web:</strong>
                    GraphQL (gqlgen), REST, FastAPI, WebSockets, JWT, Vue 3,
                    Pinia, Tailwind CSS
                </p>
                <p>
                    <strong>Infrastructure:</strong>
                    Docker Compose, Nginx, Linux, PostgreSQL, S3, CI/CD (Gitea
                    Actions)
                </p>
                <p>
                    <strong>AI / ML:</strong>
                    Claude API (structured output), PyTorch, OpenCV, Tesseract
                    OCR, Optuna
                </p>
            </section>

            <section>
                <h2>Education</h2>
                <div class="w-full flex flex-col sm:flex-row gap-5">
                    <div class="flex-1">
                        <div
                            class="flex-col flex place-content-between m-auto place-items-center"
                        >
                            <h3>
                                <a href="/pdf/transcript.pdf">
                                    University of Leeds
                                </a>
                            </h3>
                            <p>BSc Computer Science with Mathematics</p>
                            <small>
                                First Class Honours · Sep 2021 – Jun 2025
                            </small>
                        </div>
                        <ul class="list-disc list-inside pt-1">
                            <li>Algorithms & Data Structures I & II</li>
                            <li>Compilers; Formal Languages & Automata</li>
                            <li>Graph Algorithms & Complexity Theory</li>
                            <li>Machine Learning, Databases, Processors</li>
                        </ul>
                    </div>
                    <div class="flex-1">
                        <div
                            class="flex-col flex place-content-between m-auto place-items-center"
                        >
                            <h3>
                                <a href="/pdf/transcript2.pdf">
                                    University of Waterloo
                                </a>
                            </h3>
                            <p>Study Abroad Year</p>
                            <small>Sep 2023 – Apr 2024</small>
                        </div>
                        <ul class="list-disc list-inside pt-1">
                            <li>Applied Cryptography</li>
                            <li>Introduction to Computer Graphics</li>
                            <li>
                                Introduction to Rings and Fields with
                                Applications
                            </li>
                            <li>Real Analysis</li>
                        </ul>
                    </div>
                </div>
            </section>

            <section>
                <h2>Projects</h2>

                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>
                            <a
                                href="https://github.com/SteveThePug/web_server"
                            >
                                web_server.git
                            </a>
                        </h4>
                    </template>
                    <template #top>
                        <strong>
                            Go, Vue, PostgreSQL, Docker, Nginx, Gitea Actions
                        </strong>
                        <small>Nov 2025 – Present</small>
                    </template>
                    <p contenteditable="true">
                        Self-hosted platform behind my website: ~20 Docker
                        Compose services on a Linux VM: a Go GraphQL
                        API, Vue 3 SPA, FastAPI, PostgreSQL, S3 storage, Gitea
                        with CI, and the apps below.
                    </p>
                    <ul contenteditable="true">
                        <li>
                            Push-to-deploy CI/CD: tests gate releases; apps roll
                            out on healthchecks.
                        </li>
                        <li>
                            Schema-declared GraphQL auth; tests fail any
                            unguarded mutation.
                        </li>
                        <li>
                            Rate limiting, request logging, self-healing
                            restarts; migrated from a Raspberry Pi to a VM.
                        </li>
                    </ul>
                </Project>

                <Project>
                    <template #left>
                        <h4>
                            <a
                                href="https://www.adam-french.co.uk/gitea/adamf/usaco"
                            >
                                usaco.git
                            </a>
                        </h4>
                    </template>
                    <template #top>
                        <strong>Go, Vue, C++, Docker, Claude API</strong>
                        <small>Aug 2026 – Present</small>
                    </template>
                    <p contenteditable="true">
                        Bash CLI and site for my USACO, LeetCode and Codeforces
                        solutions, live at
                        <a href="https://www.adam-french.co.uk/usaco/">
                            adam-french.co.uk/usaco</a
                        >.
                    </p>
                    <ul contenteditable="true">
                        <li>
                            Git as the database: Go service indexes a clone,
                            commits edits back.
                        </li>
                        <li>
                            Sandboxed runner judges untrusted C++ and Python: no
                            network, read-only root, CPU and memory limits.
                        </li>
                        <li>
                            Claude API (JSON-schema output) explains problems,
                            and reviews my code and recorded explanations.
                        </li>
                    </ul>
                </Project>
            </section>
        </div>

        <div class="no-print w-full h-20"></div>

        <div class="cv-page justify-around">
            <section>
                <h2>More Projects</h2>

                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>
                            <a
                                href="https://www.adam-french.co.uk/gitea/adamf/job-tracker"
                            >
                                job-tracker.git
                            </a>
                        </h4>
                    </template>
                    <template #top>
                        <strong>Go, GraphQL, Vue, PostgreSQL, Claude API</strong>
                        <small>Apr 2026 – Present</small>
                    </template>
                    <p contenteditable="true">
                        Application tracker with a leads inbox: pulls postings
                        from Greenhouse, Lever, Ashby, Adzuna and Reed, scores
                        them with Claude, and captures pages via a userscript.
                        Split from the main site into its own service and
                        database.
                    </p>
                </Project>
                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>
                            <a
                                href="https://www.adam-french.co.uk/gitea/adamf/rowing_stats"
                            >
                                rowing_stats.git
                            </a>
                        </h4>
                    </template>
                    <template #top>
                        <strong>Python, OpenCV, PyTorch, Tesseract, Optuna</strong>
                        <small>Jan 2026 – Oct 2026</small>
                    </template>
                    <p contenteditable="true">
                        Turns phone photos of a rowing machine's monitor into a
                        workout log: contour detection, a CNN filter and OCR
                        that votes across six preprocessings, tuned with Optuna.
                        Replaced a paid vision API, matching it on 91% of times.
                    </p>
                </Project>
                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>
                            <a href="https://github.com/SteveThePug/tour">
                                tour.git
                            </a>
                        </h4>
                    </template>
                    <template #top>
                        <strong>Rust</strong>
                        <small>Feb 2026 – Sep 2026</small>
                    </template>
                    <p contenteditable="true">
                        CLI for authoring and stepping through code tutorials as
                        file snapshots, with Git-style staging and commits.
                    </p>
                </Project>
                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>
                            <a
                                href="https://github.com/SteveThePug/rust-raytracer"
                            >
                                rust-raytracer.git
                            </a>
                        </h4>
                    </template>
                    <template #top>
                        <strong>Rust, Linear Algebra, Multithreading</strong>
                        <small>2023</small>
                    </template>
                    <p contenteditable="true">
                        Multithreaded recursive ray tracer for realistic 3D
                        rendering; emphasised algorithmic efficiency and
                        low-level memory management in Rust.
                    </p>
                </Project>
                <Project>
                    <template #left>
                        <h4>
                            <a
                                href="https://community.wolfram.com/groups/-/m/t/3210947"
                            >
                                Wolfram Summer School
                            </a>
                        </h4>
                    </template>
                    <template #top>
                        <strong>Wolfram Mathematica</strong>
                        <small>2024</small>
                    </template>
                    <p contenteditable="true">
                        Research project on
                        <a href="/automata">Mobile Automata</a>: data
                        visualisation in Mathematica, presented to the Wolfram
                        community. Groundwork for the dissertation above.
                    </p>
                </Project>
            </section>

            <section>
                <h2>Work Experience</h2>

                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>
                            <a
                                href="https://www.daisygreenfood.com/location-hollandpark"
                            >
                                Daisy Green Holland Park
                            </a>
                        </h4>
                    </template>
                    <template #top>
                        <small>Barista & Front of House</small>
                        <small>May 2026 – Present</small>
                    </template>
                    <p contenteditable="true">
                        Coffee and table service at a busy all-day café in
                        Holland Park.
                    </p>
                </Project>
                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>
                            <a href="https://www.belgravemusichall.com/">
                                Belgrave Music Hall
                            </a>
                        </h4>
                    </template>
                    <template #top>
                        <small>Bartender & Waiter</small>
                        <small>Sep 2022 – May 2026</small>
                    </template>
                    <p contenteditable="true">
                        Through university at Belgrave Music Hall and The
                        Montpelier (Rose Pubs Group), Leeds: high-volume weekend
                        service at a live-music venue.
                    </p>
                </Project>
                <Project>
                    <template #left>
                        <h4>Earlier Roles</h4>
                    </template>
                    <template #top>
                        <small>Bar, Barista & Cashier</small>
                        <small>Jun 2021 – Sep 2022</small>
                    </template>
                    <p contenteditable="true">
                        <a href="https://crownandanchoreastbourne.co.uk/">
                            The Crown and Anchor</a
                        >
                        and
                        <a href="https://www.totherisebakery.com/">
                            To The Rise Bakery</a
                        >
                        in Eastbourne (2022);
                        <a
                            href="https://www.benugo.com/sites/restaurants/bfi-riverfront/"
                        >
                            BFI Riverfront Kitchen</a
                        >
                        in London (2021).
                    </p>
                </Project>
            </section>

            <section>
                <h2>Secondary Education</h2>
                <div
                    class="flex-row flex place-content-between m-auto place-items-center"
                >
                    <h3>University Academy of Engineering Southbank</h3>
                    <small>Sep 2014 – Jun 2021</small>
                </div>
                <p>
                    <strong>A-Levels:</strong> Mathematics A*, Further
                    Mathematics A*, Physics A*
                </p>
                <p>
                    <strong>GCSEs:</strong> Mathematics 9, English Literature 9,
                    Physics 9, English Language 8, Biology 8, Chemistry 8,
                    Computer Science 7
                </p>
            </section>

            <section>
                <h2>Interests</h2>
                <p>
                    Competitive programming (USACO, LeetCode, Codeforces) ·
                    Learning Mandarin · Rowing, climbing and the gym · Rhythm
                    games · Board games and Meetup.com
                </p>
            </section>
        </div>

        <div class="no-print w-full h-20"></div>
    </main>
</template>
