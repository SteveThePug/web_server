<script setup>
import Project from "./Project.vue";
</script>

<template>
  <main>
    <div class="no-print w-full h-20"></div>
    <div class="a4page place-content-between">
      <section>
        <div class="flex flex-col sm:flex-row sm:justify-between">
          <h1 class="content-center name">Adam French</h1>
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
          Computer Science with Mathematics graduate (First Class, 81.1%) from
          the University of Leeds, with a year abroad at the University of
          Waterloo. Three years of part-time hospitality work done alongside my
          degree. I am equally comfortable with practical, hands-on work and
          tasks that require careful thinking and I prefer roles where I am in
          direct contact with people. I am committed to building a long-term
          reliable career.
        </p>
      </section>

      <section>
        <div class="skills-grid">
          <div>
            <h2>Analytical</h2>
            <ul>
              <li>Strong with studying</li>
              <li>Systematic approach to problems</li>
              <li>Comfortable reading technical material</li>
            </ul>
          </div>
          <div>
            <h2>Practical</h2>
            <ul>
              <li>Happy with physical, hands-on tasks</li>
              <li>Patient and careful with procedures</li>
              <li>Used to checking my own work</li>
            </ul>
          </div>
          <div class="inline-skills">
            <h2>Additional</h2>
            <ul>
              <li>Three years of customer-facing work</li>
              <li>Full UK driving licence</li>
              <li>Full right to work in the UK</li>
              <li>A*A*A* A-Levels</li>
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
            <small>Bartender & Front of House</small>
            <small>2021–2025</small>
          </template>
          <p>
            Served food and drinks in a busy live-music venue in Leeds. Quick on
            feet, calm and collected through busy evenings. Worked alongside my
            degree.
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
            Front-of-house at a busy pub. Took orders, pulled pints, and kept
            service running smoothly.
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
            Prepared baked goods, operated the coffee machine, and kept the
            workspace clean and tidy.
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
            Operated the till and served customers at a café on London's South
            Bank.
          </p>
        </Project>
      </section>
      <section>
        <h2>Projects</h2>

        <Project class="border-b border-dotted">
          <template #left>
            <h4>
              <a href="https://www.adam-french.co.uk"> Personal Website </a>
            </h4>
          </template>
          <template #top>
            <small>Self-directed</small>
            <small>2023–2025</small>
          </template>
          <p>
            Built and maintained a personal website and the server it runs on,
            entirely independently. Created the initial design through to
            deployment.
          </p>
        </Project>
        <Project>
          <template #left>
            <h4>
              <a href="https://community.wolfram.com/groups/-/m/t/3210947">
                Wolfram Summer School
              </a>
            </h4>
          </template>
          <template #top>
            <small>Academic Research</small>
            <small>2024</small>
          </template>
          <p>
            Research project completed on a tight deadline in collaboration with
            academic mentors. Wrote up findings and presented to a group.
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
        </section>
        <section class="flex-1">
          <h2>University of Waterloo</h2>
          <div
            class="flex-row flex place-content-between m-auto place-items-center"
          >
            <small>Year abroad</small>
            <small>Sep 2023 – Apr 2024</small>
          </div>
        </section>
      </div>
    </div>
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
  --font-size-name: 3.5em;
  --font-size-text: 0.94em;
  --font-size-small: 0.9em;
  --font-size-heading: 1.9em;
  --font-size-subheading: 1.6em;
  --font-size-subsubheading: 1.3em;
}

.name {
  font-size: var(--font-size-name);
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
  font-size: var(--font-size-text);
  margin: 0;
  padding-left: 1.2em;
}

li {
  font-size: var(--font-size-text);
  color: var(--primary);
  list-style: disc;
}

.skills-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 0.3em 1em;
  margin-bottom: 0.2em;
}

/* Inline skills: header flows directly into bullet list on same line */
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
