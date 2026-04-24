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
          Computer Science with Mathematics graduate (First Class, 81.1%)
          applying for an electrical apprenticeship. Three cumulative years of
          part-time bar and waiting work done alongside my degree. Throughout my
          hospitality work I valued working with my hands rather than behind a
          desk. I want to earn an NVQ Level 3 in Electrical Installation to
          build a long-term career in a skilled trade.
        </p>
      </section>

      <section>
        <div class="skills-grid">
          <div>
            <h2>Technical Skills</h2>
            <ul class="list-disc list-inside">
              <li>
                Strong with calculations and digital tools, developed by my
                degree.
              </li>
              <li>
                Comfortable working and planning independently, as well as in
                teams.
              </li>
            </ul>
          </div>
          <div>
            <h2>Practical Skills</h2>
            <ul class="list-disc list-inside">
              <li>Careful and patient with procedures.</li>
              <li>Used to checking own work thoroughly.</li>
            </ul>
          </div>
          <div>
            <h2>Additional</h2>
            <ul class="list-disc list-inside">
              <li>A*A*A* A-Levels.</li>
              <li>Full UK driving licence.</li>
              <li>Full right to work in the UK.</li>
              <li>Eager to complete training to a high standard.</li>
              <li>References on request.</li>
            </ul>
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
            Served food and drinks in a high-volume live-music venue in Leeds.
            Handled busy weekend shifts, managed multiple tables simultaneously,
            and maintained a calm, friendly demeanour during peak hours.
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
          <p>
            Worked front-of-house at a busy pub, pulling pints, taking orders,
            and ensuring a welcoming atmosphere. Built rapport with regulars and
            adapted quickly to changing priorities during service.
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
          <p>
            Worked at a bakery in Eastbourne, preparing freshly baked goods for
            display, operating the coffee machine, and keeping bakery equipment
            spotless. Built rapport with regulars and provided attentive,
            friendly service.
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
          <p>
            Operated the till, served customers, and helped coordinate table
            service at a café on London's South Bank. Developed strong
            cash-handling accuracy and customer interaction skills in a
            fast-paced environment.
          </p>
        </Project>
      </section>

      <div class="w-full flex flex-col sm:flex-row gap-5">
        <section class="flex-1">
          <h2>
            <a href="https://www.adam-french.co.uk/pdf/transcript.pdf">
              University of Leeds
            </a>
          </h2>
          <div
            class="flex-row flex place-content-between m-auto place-items-center"
          >
            <small>First Class Honours (81.1%)</small>
            <small>Sep 2021 – Jun 2025</small>
          </div>
          <small>BSc Computer Science with Mathematics</small>
          <ul class="list-disc list-inside">
            <li>Algorithms & Data Structures I & II</li>
            <li>Compiler Design and Construction</li>
            <li>Formal Languages & Finite Automata</li>
            <li>Graph Algorithms & Complexity Theory</li>
            <li>Machine Learning, Databases, Computer Processors</li>
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
            <li>Introduction to Rings and Fields with Applications</li>
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
