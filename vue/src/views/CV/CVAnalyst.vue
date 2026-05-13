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
        <p contenteditable="true">
          Graduate with a First Class Honours degree in Computer Science with
          Mathematics from the University of Leeds (81.1%) and a year abroad at
          the University of Waterloo. Strong foundations in statistics, machine
          learning, and databases, paired with practical experience turning raw
          data into clear, actionable insight. Comfortable across the analytics
          stack, from SQL queries and Python notebooks to written reports.
          Looking to contribute rigorous, well-communicated analysis to a
          data-driven team.
        </p>
      </section>

      <section>
        <h2>Skills</h2>
        <div class="skills-grid">
          <div>
            <strong>Analysis & Visualisation</strong><br /><small
              >Python (Pandas, NumPy, scikit-learn), matplotlib, seaborn,
              Wolfram Mathematica</small
            >
          </div>
          <div>
            <strong>Statistics & ML</strong><br /><small
              >Regression, hypothesis testing, classification, clustering,
              A/B testing</small
            >
          </div>
          <div>
            <strong>Databases & Tooling</strong><br /><small
              >SQL, PostgreSQL, SQLite, Jupyter, Git, Docker</small
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
                href="https://community.wolfram.com/groups/-/m/t/3210947"
              >
                Wolfram Summer School
              </a>
            </h4>
          </template>
          <template v-slot:top>
            <small>Wolfram Mathematica, Data Visualisation</small>
            <small>2024</small>
          </template>
          <p contenteditable="true">
            Research on Mobile Automata: simulated rule-based agents, analysed
            long-run behaviour, and built visualisations. Delivered a write-up
            and live presentation within a two-week deadline.
          </p>
        </Project>

        <Project class="border-b border-dotted">
          <template v-slot:left>
            <h4>
              <a
                href="https://www.adam-french.co.uk/gitea/adamf/rowing_stats"
              >
                rowing_stats.git
              </a>
            </h4>
          </template>
          <template v-slot:top>
            <small>Python, scikit-learn, Pandas</small>
            <small>2024</small>
          </template>
          <p contenteditable="true">
            Extracted workout data from Concept 2 rowing machines and ran
            regression and ML analysis on the resulting dataset to model
            performance trends over time.
          </p>
        </Project>

        <Project class="border-b border-dotted">
          <template v-slot:left>
            <h4>
              <a href="https://www.adam-french.co.uk/gitea/adamf/tour">
                tour.git
              </a>
            </h4>
          </template>
          <template v-slot:top>
            <small>Rust</small>
            <small>2026</small>
          </template>
          <p contenteditable="true">
            CLI tool for building and navigating interactive code tutorials,
            with Git-inspired version traversal. Demonstrates careful planning
            and code structuring.
          </p>
        </Project>

        <Project>
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
            <small>Postgres, Go, Python, Docker</small>
            <small>2025</small>
          </template>
          <p contenteditable="true">
            Self-hosted personal website with a Postgres-backed API. Wrote
            queries and scripts to inspect user activity and content data.
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
            <li>Machine Learning</li>
            <li>Databases</li>
            <li>Probability, Statistics & Linear Algebra</li>
            <li>Algorithms & Data Structures I & II</li>
            <li>Graph Algorithms & Complexity Theory</li>
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
            <h4>Hospitality</h4>
          </template>
          <template #top>
            <small>Cashier, Bartender, Waiter</small>
            <small>Jan 2018 – Dec 2023</small>
          </template>
          <p contenteditable="true">
            Worked at <strong>Belgrave Music Hall</strong>,
            <strong>The Crown and Anchor</strong>, and
            <strong>BFI Riverfront Kitchen</strong>. Developed clear
            communication, composure under pressure, and reliability.
          </p>
        </Project>
      </section>
      <div class="w-full flex flex-col sm:flex-row gap-5">
        <section class="flex-1">
          <h2>Soft Skills</h2>
          <ul class="list-disc list-inside">
            <li>Clear written & verbal communication</li>
            <li>Attention to detail</li>
            <li>Structured problem solving</li>
            <li>Collaboration</li>
            <li>Time management under deadlines</li>
          </ul>
        </section>
        <section class="flex-1">
          <h2>Interests</h2>
          <ul class="list-disc list-inside">
            <li>Leetcode & data puzzles</li>
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
