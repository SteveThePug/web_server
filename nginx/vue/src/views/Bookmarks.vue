<script setup>
import { ref } from "vue";

import ToggleLinkTable from "@/components/quick/ToggleLinkTable.vue";

const job_links = [
    {
        name: "LinkedIn",
        link: "https://www.linkedin.com/",
    },
    {
        name: "Jack and Jill",
        link: "https://app.jackandjill.ai",
    },
    {
        name: "LinkedIn",
        link: "https://www.linkedin.com/",
    },
    {
        name: "Prospects",
        link: "https://www.prospects.ac.uk/",
    },
    {
        name: "GOV",
        link: "https://findajob.dwp.gov.uk",
    },
    {
        name: "Glassdoor",
        link: "https://www.glassdoor.co.uk/",
    },
    {
        name: "Indeed",
        link: "https://www.indeed.co.uk/",
    },
];

const learning_links = [
    {
        name: "Leetcode",
        link: "https://leetcode.com/",
    },
    {
        name: "ISLP",
        link: "https://hastie.su.domains/ISLP/ISLP_website.pdf.download.html",
    },
];

const social_links = [
    {
        name: "Outlook",
        link: "https://outlook.live.com/",
    },
    {
        name: "Gmail",
        link: "https://mail.google.com/",
    },
    {
        name: "Whatsapp",
        link: "https://web.whatsapp.com/",
    },
];

const radio_links = [
    {
        name: "Radio Helsinki",
        link: "https://www.radiohelsinki.fi/",
    },
    {
        name: "Palanga Street Radio",
        link: "https://palanga.live/",
    },
    {
        name: "IDA Radio",
        link: "https://idaidaida.net/",
    },
    {
        name: "Tīrkultūra",
        link: "https://www.tirkultura.lv/",
    },
];

const hacking_links = [
    {
        name: "pwn.college",
        link: "https://pwn.college/",
    },
    {
        name: "OSINT Framework",
        link: "https://osintframework.com/",
    },
    {
        name: "OverTheWire",
        link: "https://overthewire.org/",
    },
    {
        name: "TryHackMe",
        link: "https://tryhackme.com/",
    },
];

const chinese_links = [
    {
        name: "MDBG Chinese Dictionary",
        link: "https://www.mdbg.net/chinese/dictionary",
    },
    {
        name: "Stroke Order",
        link: "https://www.strokeorder.com/",
    },
    {
        name: "HSK 1 Peking University",
        link: "https://youtube.com/playlist?list=PLVWfp7qXLmKVfSUkucXErLncKn-JqgBbK&si=2ytO3inS8-iOAOx2",
    },
    {
        name: "Stroke Order",
        link: "https://www.strokeorder.com/",
    },
    {
        name: "Offbeat Mandarin",
        link: "https://www.youtube.com/@OffbeatMandarin",
    },
];

const art_links = [
    {
        name: "Frida Kahlo",
        link: "https://www.fridakahlo.org/",
    },
    {
        name: "Cameron's World",
        link: "https://www.cameronsworld.net/",
    },
    {
        name: "Neocities",
        link: "https://neocities.org/",
    },
];

const vue_links = [
    {
        name: "Vue",
        link: "https://vuejs.org/guide/introduction.html",
    },
    {
        name: "Vue Router",
        link: "https://router.vuejs.org/introduction.html",
    },
    {
        name: "Pinia",
        link: "https://pinia.vuejs.org/introduction.html",
    },
];

const go_links = [
    {
        name: "Golang",
        link: "https://golang.org/doc/",
    },
    {
        name: "Gin Gonic",
        link: "https://gin-gonic.com/en/docs/introduction/",
    },
    {
        name: "GORM",
        link: "https://gorm.io/gen/index.html",
    },
];

const doc_links = [
    {
        name: "Rust",
        link: "https://doc.rust-lang.org/stable/book/index.html",
    },
    {
        name: "Javascript",
        link: "https://developer.mozilla.org/en-US/docs/Web/JavaScript",
    },
    {
        name: "Python",
        link: "https://docs.python.org/3/",
    },
];
</script>

<template>
    <main class="center-content flex-col">
        <div class="background halftone" />
        <div class="a4page-portrait bdr-1 flex-col relative scroll-y gap">
            <ToggleLinkTable title="Learning" :linkArr="learning_links" />
            <ToggleLinkTable title="Radio" :linkArr="radio_links" />
            <ToggleLinkTable title="Hacking" :linkArr="hacking_links" />
            <ToggleLinkTable title="Chinese" :linkArr="chinese_links" />
            <ToggleLinkTable title="Web Art" :linkArr="art_links" />
            <ToggleLinkTable title="Go Docs" :linkArr="go_links" />
            <ToggleLinkTable title="Vue Docs" :linkArr="vue_links" />
            <ToggleLinkTable title="Job Apps" :linkArr="job_links" />
            <ToggleLinkTable title="Social" :linkArr="social_links" />
            <ToggleLinkTable title="Other Docs" :linkArr="doc_links" />
        </div>
    </main>
</template>
