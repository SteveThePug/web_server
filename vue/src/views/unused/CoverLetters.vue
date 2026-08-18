<script setup>
import { ref } from "vue";

const name = ref("Adam French");
const phone = ref("+447563266931");
const email = ref("adam.a.french@outlook.com");
const role = ref("Job title / role");
const paragraphs = ref([
    "Why I'm interested in the company...",
    "Why I'm a strong fit for the role...",
    "Relevant experience and background...",
]);
const signoff = ref("Thank you for reading - Adam F");

function addParagraph() {
    paragraphs.value.push("New paragraph...");
}

function removeParagraph(index) {
    paragraphs.value.splice(index, 1);
}

function updateParagraph(index, event) {
    paragraphs.value[index] = event.target.innerText;
}
</script>

<template>
    <main>
        <div class="a5page">
            <div class="contact">
                <h1
                    contenteditable="true"
                    @blur="name = $event.target.innerText"
                >
                    {{ name }}
                </h1>
                <div class="contact-details">
                    <p
                        contenteditable="true"
                        @blur="phone = $event.target.innerText"
                    >
                        {{ phone }}
                    </p>
                    <p
                        contenteditable="true"
                        @blur="email = $event.target.innerText"
                    >
                        {{ email }}
                    </p>
                </div>
            </div>
            <h2 contenteditable="true" @blur="role = $event.target.innerText">
                {{ role }}
            </h2>
            <div
                v-for="(p, i) in paragraphs"
                :key="i"
                class="paragraph-wrapper"
            >
                <p contenteditable="true" @blur="updateParagraph(i, $event)">
                    {{ p }}
                </p>
                <button
                    class="no-print remove-btn"
                    title="Remove paragraph"
                    @click="removeParagraph(i)"
                >
                    ×
                </button>
            </div>
            <p
                contenteditable="true"
                @blur="signoff = $event.target.innerText"
            >
                {{ signoff }}
            </p>
            <button class="no-print add-btn" @click="addParagraph()">
                + Add paragraph
            </button>
        </div>
    </main>
</template>

<style scoped>
.a5page {
    max-width: 210mm;
    min-height: 148mm;
    margin: 1rem auto;
    padding: 2rem 3rem;
    background: white;
    color: #222;
    line-height: 1.5;
}

.a5page h1 {
    font-size: 1.6rem;
    font-weight: 700;
}

.a5page h2 {
    font-size: 1.1rem;
    font-weight: 600;
    margin: 0.75rem 0;
}

.a5page p {
    margin-bottom: 0.75rem;
}

.contact {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    border-bottom: 1px solid #333;
    padding-bottom: 0.5rem;
    margin-bottom: 0.75rem;
}

.contact-details {
    text-align: right;
    font-size: 0.85rem;
}

.contact-details p {
    margin: 0;
}

[contenteditable="true"]:hover,
[contenteditable="true"]:focus {
    outline: 1px dashed #999;
    outline-offset: 2px;
}

.paragraph-wrapper {
    position: relative;
}

.paragraph-wrapper .remove-btn {
    position: absolute;
    top: 0;
    right: -1.5rem;
    border: none;
    background: none;
    color: #c00;
    cursor: pointer;
    font-size: 1rem;
    opacity: 0;
}

.paragraph-wrapper:hover .remove-btn {
    opacity: 1;
}

.add-btn {
    margin-top: 0.5rem;
    padding: 0.3rem 0.8rem;
    border: 1px solid #333;
    border-radius: 4px;
    background: white;
    color: #333;
    cursor: pointer;
    font-size: 0.85rem;
}

.add-btn:hover {
    background: #eee;
}

@media print {
    @page {
        size: A5 landscape;
        margin: 0;
    }
    .no-print {
        display: none !important;
    }
}
</style>
