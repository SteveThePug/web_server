<script setup>
import { ref } from "vue";

import ToggleLinkTable from "@/components/util/ToggleLinkTable.vue";

const links = [
    [
        "Reading Links",
        [
            {
                name: "Substack",
                link: "https://substack.com/",
            },
            {
                name: "Medium",
                link: "https://medium.com/",
            },
            {
                name: "4Chan",
                link: "https://www.4chan.org/",
            },
        ],
    ],
    [
        "Job Links",
        [
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
        ],
    ],
    [
        "Learning Links",
        [
            {
                name: "Leetcode",
                link: "https://leetcode.com/",
            },
            {
                name: "ISLP",
                link: "https://hastie.su.domains/ISLP/ISLP_website.pdf.download.html",
            },
        ],
    ],
    [
        "Social Links",
        [
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
        ],
    ],
    [
        "Radio links",
        [
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
        ],
    ],
    [
        "Hacking Links",
        [
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
        ],
    ],
    [
        "Chinese Links",
        [
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
        ],
    ],
    [
        "Art links",
        [
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
        ],
    ],
    [
        "Vue links",
        [
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
        ],
    ],
    [
        "Go links",
        [
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
        ],
    ],
    [
        "Doc links",
        [
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
        ],
    ],
    [
        "Article links",
        [
            {
                name: "Go and GORM",
                link: "https://medium.com/@chaewonkong/learn-go-understanding-and-implementing-foreign-keys-with-gorm-6d7608e1dbf6",
            },
            {
                name: "JWT Auth in GO",
                link: "https://medium.com/monstar-lab-bangladesh-engineering/jwt-auth-in-go-dde432440924",
            },
            {
                name: "Websockets in GO",
                link: "https://medium.com/@tanngontn/golang-gin-framework-with-normal-websocket-and-websocket-with-producer-is-rabbitmq-guide-93cad7d290f7",
            },
        ],
    ],
];
</script>

<template>
    <main class="items-center flex flex-col halftone">
        <div
            class="a4page-portrait bdr-1 flex flex-row flex-wrap overflow-x-auto gap-1"
        >
            <div class="w-full h-fit">
                <ToggleLinkTable
                    class="flex flex-col flex-wrap"
                    v-for="link in links"
                    :title="link[0]"
                    :linkArr="link[1]"
                />
            </div>
        </div>
    </main>
</template>
