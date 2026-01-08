<script setup>
import LinkTable from "@/components/quick/LinkTable.vue";

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
];
</script>

<template>
    <main class="center-content flex-col">
        <div class="a4page-portrait bdr-1 flex-col relative scroll-y gap">
            <h1>Radio Links</h1>
            <LinkTable :linkArr="radio_links" />
            <span>---</span>
            <h1>Hack</h1>
            <LinkTable :linkArr="hacking_links" />
            <span>---</span>
            <h1>Chinese</h1>
            <LinkTable :linkArr="chinese_links" />
            <span>---</span>
            <h1>Art</h1>
            <LinkTable :linkArr="art_links" />
        </div>
    </main>
</template>
