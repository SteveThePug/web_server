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

      <section>
        <h2>Profile</h2>
        <p>
          First Class Honours graduate in Computer Science with Mathematics from
          the University of Leeds (81.1%), with a year abroad at the University
          of Waterloo. Passionate about crafting responsive, accessible, and
          performant user interfaces. Experienced across multiple frontend
          frameworks with a solid understanding of the full stack.
        </p>
      </section>

      <section>
        <h2>Skills</h2>
        <div class="skills-grid">
          <div>
            <strong>Frontend</strong><br /><small
              >Vue, React / Redux, Svelte, Tailwind CSS, HTML / CSS,
              WebAssembly</small
            >
          </div>
          <div>
            <strong>Languages</strong><br /><small
              >JavaScript / TypeScript, Rust, Go, Python, SQL</small
            >
          </div>
          <div>
            <strong>Tooling / Infra</strong><br /><small
              >Vite, Docker, Nginx, Git Actions, PostgreSQL, Figma</small
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
              Vue 3, Tailwind CSS, Vite, Pinia, Responsive Design, Rust → Wasm
            </small>
            <small>2025</small>
          </template>
          <p>
            Personal website SPA built with Vue 3, Tailwind CSS, and Pinia for
            state management. Features responsive layouts, dark mode,
            WebAssembly integration, and a custom component library. Iterated
            through Svelte and React/Redux before settling on Vue.
          </p>
        </Project>
        <Project class="border-b border-dotted">
          <template v-slot:left>
            <h4>
              <a href="https://www.adam-french.co.uk/gitea/adamf/tour.git">
                tour.git
              </a>
            </h4>
          </template>
          <template v-slot:top>
            <small>Rust</small>
            <small>2026</small>
          </template>
          <p>
            CLI tool for building and navigating interactive code tutorials,
            with version-traversal semantics inspired by Git.
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
            Parallelised recursive ray tracer for realistic 3D rendering.
            Emphasised algorithmic efficiency and low-level memory management in
            Rust.
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
            Research project on Mobile Automata with data visualisation and
            academic presentation. Delivered within a tight deadline in
            collaboration with academic mentors.
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
          <small>BSc Computer Science with Mathematics </small>
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
            <em>The Crown and Anchor</em>, and <em>BFI Riverfront Kitchen</em>.
            Developed communication, composure under pressure, and reliability
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
