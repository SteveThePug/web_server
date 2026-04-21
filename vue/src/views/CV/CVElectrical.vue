<script setup>
import Project from "./Project.vue";
</script>

<template>
    <main>
        <div class="no-print w-full h-20"></div>
        <div class="a4page justify-between">
            <section>
                <div class="flex flex-col sm:flex-row sm:justify-between">
                    <h1 class="name">Adam French</h1>
                    <div class="contact-details sm:text-right">
                        <p>London, United Kingdom</p>
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
                    Computer Science with Mathematics graduate (First
                    Class, 81.1%) applying for an electrical
                    apprenticeship. Five years of bar and waiting work
                    alongside my degree, and a background of building and
                    running my own home server setup. After finishing
                    university I've decided I want to work with my hands
                    rather than behind a desk, and I'm ready to start at
                    the bottom and work through the NVQ.
                </p>
            </section>

            <section>
                <h2>Key Strengths</h2>
                <div class="skills-grid">
                    <div>
                        <strong>Technical</strong><br /><small>
                            Good with numbers from my Maths degree.
                            Comfortable with Excel and the Microsoft
                            Suite. Pick up new tools quickly.
                        </small>
                    </div>
                    <div>
                        <strong>Practical</strong><br /><small>
                            Careful and patient with procedure. Used to
                            checking my own work before calling it done.
                            Take safety rules seriously.
                        </small>
                    </div>
                    <div>
                        <strong>People</strong><br /><small>
                            Five years of customer-facing work. Reliable
                            turning up, easy to work with, keep my head
                            when a shift gets busy.
                        </small>
                    </div>
                </div>
            </section>

            <section>
                <h2>Experience</h2>

                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>Belgrave Music Hall</h4>
                    </template>
                    <template #top>
                        <small>Bartender & Waiter</small>
                        <small>2021–2025</small>
                    </template>
                    <p>
                        Busy live-music venue in Leeds. Worked the bar
                        and floor across gig nights and event days,
                        following the usual health, safety and hygiene
                        procedures. Got on well with the team and could
                        be trusted to open or close on my own.
                    </p>
                </Project>
                <Project class="border-b border-dotted">
                    <template #left>
                        <h4>BFI Riverfront Kitchen</h4>
                    </template>
                    <template #top>
                        <small>Cashier & Waiter</small>
                        <small>2018–2020</small>
                    </template>
                    <p>
                        Ran the till and waited tables at a café on the
                        South Bank. First proper job. Learned how to keep
                        cash-handling tidy and how to talk to a lot of
                        different people without getting flustered.
                    </p>
                </Project>
                <Project>
                    <template #left>
                        <h4>Personal Projects</h4>
                    </template>
                    <template #top>
                        <small>Self-directed learning</small>
                        <small>2021–present</small>
                    </template>
                    <p>
                        Built my own website and host it at home on a
                        Raspberry Pi. Set up the networking and low-voltage
                        DC wiring myself. When something breaks I work
                        through it until it's fixed, which is the same
                        mindset I'd bring to a site.
                    </p>
                </Project>
            </section>

            <section>
                <h2>Education</h2>
                <div class="w-full flex flex-col sm:flex-row gap-5">
                    <div class="flex-1 sm:border-r border-dotted sm:pr-3">
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
                        <small>BSc Computer Science with Mathematics</small>
                    </div>
                    <div class="flex-1 sm:pl-3">
                        <h3>GCSEs</h3>
                        <small>
                            Including English Language and Mathematics,
                            both well above the Grade 4/5 (C) requirement.
                        </small>
                    </div>
                </div>
            </section>

            <section>
                <h2>Additional</h2>
                <ul class="list-disc list-inside">
                    <li>
                        Full right to work in the UK. Happy to relocate or
                        travel for the right role.
                    </li>
                    <li>
                        Fine with a safety-critical medical as part of
                        recruitment.
                    </li>
                    <li>
                        Up for college day-release and seeing the NVQ all
                        the way through.
                    </li>
                    <li>References on request.</li>
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
    --font-size-subheading: 1.6em;
    --font-size-subsubheading: 1.3em;
}

/* A4 Page */
.a4page {
    line-height: 1.5;
    font-family: var(--font-text);
    width: 210mm;
    height: 297mm;
    padding: 8mm;
    box-sizing: border-box;
    background-color: var(--background);
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
    border: 1px solid var(--primary);
    overflow: hidden;
    margin: auto auto;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    gap: 0.4em;
}

/* Component Styling */
main {
    padding: 0px;
    display: flex;
    flex-direction: column;
    height: fit-content;
    background-color: white;
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
}

p {
    margin-bottom: 0.2em;
    color: var(--primary);
    font-size: var(--font-size-text);
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
