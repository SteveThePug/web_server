<script setup>
import Project from "./Project.vue";
</script>

<template>
  <main class="cv-template">
    <div class="no-print w-full h-20"></div>
    <div class="a4page justify-between">
      <section class="flex flex-col sm:flex-row sm:justify-between">
        <h1 class="name">Adam French</h1>
        <div class="contact-details">
          <p>London, United Kingdom</p>
          <p>+447563266931</p>
          <p>adam.a.french@outlook.com</p>
          <p>
            <a href="https://www.adam-french.co.uk"> www.adam-french.co.uk </a>
          </p>
        </div>
      </section>

      <section>
        <h2>Profile</h2>
        <p>
          First Class Honours graduate in Computer Science with Mathematics from
          the University of Leeds (81.1%), with a year abroad at the University
          of Waterloo. Passionate about developer productivity, automation
          infrastructure, and software testing at scale. Experienced building
          CI/CD pipelines, automation tooling, and scalable backend services.
          Eager to apply rigorous engineering discipline within a collaborative
          platform team.
        </p>
      </section>

      <section>
        <h2>Skills</h2>
        <div class="skills-grid">
          <div>
            <strong>Languages</strong><br /><small
              >Python, Go, Rust, Swift, JavaScript / TypeScript, SQL</small
            >
          </div>
          <div>
            <strong>Automation &amp; Testing</strong><br /><small
              >CI/CD Pipelines, GitHub Actions, Docker, Unit &amp; Integration
              Testing, WebAssembly</small
            >
          </div>
          <div>
            <strong>Infrastructure</strong><br /><small
              >Nginx, PostgreSQL, SQLite, JWT Auth, REST &amp; GraphQL
              APIs</small
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
              GitHub Actions, Docker, Nginx, Go, Python, Rust → Wasm, Postgres,
              JWT Auth
            </small>
            <small>2025</small>
          </template>
          <p>
            Self-hosted personal website with a fully automated CI/CD pipeline:
            lint, build, test, and deploy on every push. Designed for
            zero-downtime deployments on constrained Raspberry Pi hardware.
            Iterated across diverse stacks to evaluate tradeoffs in
            infrastructure and developer experience.
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
            <small>Rust, Developer Tooling, CLI</small>
            <small>2026</small>
          </template>
          <p>
            Developer productivity CLI for building and navigating interactive
            code tutorials with Git-inspired version traversal. Designed as a
            reusable automation library with a clean API surface for embedding
            in larger toolchains.
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
    </div>

    <div class="no-print w-full h-20"></div>

    <div class="a4page gap-10">
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
        <h2>Interests</h2>
        <ul class="list-disc list-inside">
          <li>Leetcode — daily competitive problem solving</li>
          <li>iOS/macOS development — personal Swift projects</li>
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
