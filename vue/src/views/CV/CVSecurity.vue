<script setup>
import Project from "./Project.vue";
</script>

<template>
  <main class="cv-template">
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
          Computer Science with Mathematics graduate (First Class Honors, 81.1%)
          from the University of Leeds, with a year abroad at the University of
          Waterloo. Hands-on experience self-hosting and hardening web
          infrastructure, implementing authentication systems, and writing
          memory-safe systems code. Methodical, detail-driven, and enjoy
          adversarial problem-solving.
        </p>
      </section>

      <section>
        <h2>Skills</h2>
        <div class="skills-grid">
          <div>
            <strong>Security</strong><br /><small
              >Applied Cryptography, JWT Auth, TLS / Certbot, Rate-Limiting,
              OWASP Top 10</small
            >
          </div>
          <div>
            <strong>Languages & Systems</strong><br /><small
              >Rust, Go, Python, C, JavaScript / TypeScript, SQL</small
            >
          </div>
          <div>
            <strong>Infra & Tooling</strong><br /><small
              >Linux, Nginx, Docker, PostgreSQL, Git, CI/CD, WebAssembly</small
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
              Nginx, Docker, JWT Auth, TLS, Rate-Limiting, Go, Postgres
            </small>
            <small>2025</small>
          </template>
          <p contenteditable="true">
            Self-hosted personal site running on a Raspberry Pi. Implemented JWT
            auth with HTTP-only cookies, Nginx rate-limiting on login and
            uploads, automated TLS via Certbot, and a hardened reverse-proxy
            configuration.
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
            <small>Rust, Multithreading, Memory Safety</small>
            <small>2023</small>
          </template>
          <p contenteditable="true">
            Parallelised recursive ray tracer written in Rust, emphasising
            low-level memory management, safe concurrency, and disciplined
            handling of unsafe boundaries.
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
            <small>Rust, CLI, Version Control Internals</small>
            <small>2026</small>
          </template>
          <p contenteditable="true">
            CLI tool for building and navigating interactive code tutorials,
            with version-traversal semantics inspired by Git.
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
            <small>First Class Honors (81.1%)</small>
            <small>Sep 2021 – Jun 2025</small>
          </div>
          <small>BSc Computer Science with Mathematics </small>
          <ul class="list-disc list-inside">
            <li>Algorithms & Data Structures I & II</li>
            <li>Compiler Design and Construction</li>
            <li>Formal Languages & Finite Automata</li>
            <li>Graph Algorithms & Complexity Theory</li>
            <li>Computer Processors, Databases, Networks</li>
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
            <li>Introduction to Rings and Fields with Applications</li>
            <li>Introduction to Computer Graphics</li>
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
            <strong>BFI Riverfront Kitchen</strong>. Developed communication,
            composure under pressure, and reliability in customer-facing roles.
          </p>
        </Project>
      </section>
      <div class="w-full flex flex-col sm:flex-row gap-5">
        <section class="flex-1">
          <h2>Soft Skills</h2>
          <ul class="list-disc list-inside">
            <li>Methodical, evidence-led problem solving</li>
            <li>Attention to detail</li>
            <li>Clear written communication</li>
            <li>Self-directed learning</li>
            <li>Composure under pressure</li>
          </ul>
        </section>
        <section class="flex-1">
          <h2>Interests</h2>
          <ul class="list-disc list-inside">
            <li>Capture The Flag (CTF) challenges</li>
            <li>Leetcode</li>
            <li>Learning Mandarin</li>
            <li>Climbing, Gym</li>
            <li>Board games, Meetup.com</li>
          </ul>
        </section>
      </div>
    </div>

    <div class="no-print w-full h-20"></div>
  </main>
</template>
