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
          the University of Leeds (81.1%). Dependable and personable team player
          with five years of hospitality experience across busy bars,
          restaurants, and event venues. Thrives under pressure, communicates
          clearly, and takes pride in providing excellent customer service.
        </p>
      </section>

      <section>
        <h2>Experience</h2>

        <Project class="border-b border-dotted">
          <template #left>
            <h4>
              <a href="https://www.daisygreenfood.com/location-hollandpark">
                Daisy Green Holland Park
              </a>
            </h4>
          </template>
          <template #top>
            <small>Barista & Front of House</small>
            <small>May 2026 – Present</small>
          </template>
          <p>
            Working at an all-day café in Holland Park, preparing coffee and
            serving food in a fast-paced environment.
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

      <section>
        <h2>Skills</h2>
        <div class="skills-grid">
          <div>
            <strong>Service</strong><br /><small
              >Bar work, Table service, Cash handling, Till operation, Food
              hygiene</small
            >
          </div>
          <div>
            <strong>Soft Skills</strong><br /><small
              >Communication, Teamwork, Time management, Composure under
              pressure</small
            >
          </div>
          <div>
            <strong>Technical</strong><br /><small
              >EPOS systems, Stock management, Event coordination</small
            >
          </div>
        </div>
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
            <small>2021–2025</small>
          </div>
          <small>BSc Computer Science with Mathematics </small>
        </section>
        <section class="flex-1">
          <h2>University of Waterloo</h2>
          <div
            class="flex-row flex place-content-between m-auto place-items-center"
          >
            <small>Year abroad</small>
            <small>2023–2024</small>
          </div>
        </section>
      </div>

      <section>
        <h2>Interests</h2>
        <ul class="list-disc list-inside">
          <li>Leetcode</li>
          <li>Rhythm Games</li>
          <li>Climbing, Gym</li>
          <li>Board games, Meetup.com</li>
        </ul>
      </section>
    </div>

    <div class="no-print w-full h-20"></div>
  </main>
</template>
