<script setup>
import Header from "@/components/text/Header.vue";
import Paragraph from "@/components/text/Paragraph.vue";
import HostOverview from "../diagrams/infra/HostOverview.vue";
</script>

<template>
    <section id="infra" class="flex flex-col gap-6">
        <Header>Infrastructure &amp; deployment</Header>
        <Paragraph>
            The whole site is one Docker Compose stack on a single Ubuntu VM,
            <code>stpvm</code>. Only nginx (and the Cloudflare tunnel) faces
            the internet; everything else talks over the internal Docker
            network.
        </Paragraph>
        <HostOverview />
    </section>
</template>
