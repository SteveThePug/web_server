<script setup>
import Header from "@/components/text/Header.vue";
import Paragraph from "@/components/text/Paragraph.vue";
</script>

<template>
    <div
        class="flex-1 border-box flex flex-col p-1 text-left items-start justify-start"
    >
        <Header>Intro</Header>
        <Paragraph>
            Hi, I'm Adam, thank you for visiting my website. I'm currently a 20
            something graduate looking for work. I like to game, listen to lots
            of music and occasionally watch anime.
        </Paragraph>
        <Header>Getting around</Header>
        <Paragraph>
            Pages available can be traversed through links below. I am hoping to
            add some shrines, code-walkthoughs, live chat and page transitions
            at a later date.
        </Paragraph>
        <Header>Contact</Header>
        <Paragraph>
            Please email me <a href="mailto:adam.a.french@outlook.com">here</a>,
            or contact me though any of the social medias linked.
        </Paragraph>
        <Header>A Quote</Header>
        <!-- <p>
            What makes me a good demoman? If I were a bad demoman, I wouldn't be
            sittin' here discussin' it with you, now would I?!
        </p> -->
        <Paragraph>
            One crossed wire, one wayward pinch of potassium chlorate, one
            errant twitch, and KA-BLOOIE!
        </Paragraph>
    </div>
</template>
