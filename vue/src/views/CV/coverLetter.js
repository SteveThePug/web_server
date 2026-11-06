/**
 * The cover letter as data, with no Vue in it: the default letter, the JSON
 * body stored in a CoverLetter row (stores/coverLetters.js), and the moves
 * the drag and drop in CoverLetters.vue makes.
 *
 * Body shape (v: 1):
 *   { v, name, phone, email, role, signoff,
 *     sections: [{ id, title, paragraphs: [{ id, text }] }] }
 *
 * Every section and paragraph carries an id because the template keys its
 * v-for on it. Keying on the index would leave a contenteditable showing the
 * old text after a reorder, since Vue would reuse the element in place.
 *
 * A position is `{ section, index }`: a section's index and a paragraph's
 * index inside it. `index` may equal the paragraph count, meaning "at the end".
 */

export const BODY_VERSION = 1;

/** A short id, unique enough within one letter. crypto.randomUUID would do,
 *  but it is missing outside a secure context (the dev stack over a LAN IP). */
export function newId() {
  return Math.random().toString(36).slice(2, 10) + Date.now().toString(36);
}

export function newParagraph(text = "New paragraph...") {
  return { id: newId(), text };
}

export function newSection(title = "New section", texts = []) {
  return { id: newId(), title, paragraphs: texts.map(newParagraph) };
}

/** The blank template a visitor sees, and what "New" starts from. */
export function newLetter() {
  return {
    v: BODY_VERSION,
    name: "Adam French",
    phone: "+447563266931",
    email: "adam.a.french@outlook.com",
    role: "Job title / role",
    signoff: "Thank you for reading - Adam F",
    sections: [
      newSection("Why the company", ["Why I'm interested in the company..."]),
      newSection("Why I fit", ["Why I'm a strong fit for the role..."]),
      newSection("Experience", ["Relevant experience and background..."]),
    ],
  };
}

const str = (v, fallback) => (typeof v === "string" ? v : fallback);

/**
 * Read a stored body back into a letter. Anything missing or malformed
 * falls back to the default field by field, so a hand-edited or older row
 * still opens rather than blanking the page; ids are regenerated if absent.
 */
export function parseBody(json) {
  const base = newLetter();
  let raw;
  try {
    raw = JSON.parse(json);
  } catch {
    return base;
  }
  if (!raw || typeof raw !== "object") return base;
  const sections = Array.isArray(raw.sections)
    ? raw.sections
        .filter((s) => s && typeof s === "object")
        .map((s) => ({
          id: str(s.id, newId()),
          title: str(s.title, ""),
          paragraphs: Array.isArray(s.paragraphs)
            ? s.paragraphs
                .filter((p) => p && typeof p.text === "string")
                .map((p) => ({ id: str(p.id, newId()), text: p.text }))
            : [],
        }))
    : base.sections;
  return {
    v: BODY_VERSION,
    name: str(raw.name, base.name),
    phone: str(raw.phone, base.phone),
    email: str(raw.email, base.email),
    role: str(raw.role, base.role),
    signoff: str(raw.signoff, base.signoff),
    sections,
  };
}

/** The JSON stored as a CoverLetter's body. Also what the "unsaved" marker
 *  compares, so the key order here must stay fixed. */
export function serialize(letter) {
  return JSON.stringify({
    v: BODY_VERSION,
    name: letter.name,
    phone: letter.phone,
    email: letter.email,
    role: letter.role,
    signoff: letter.signoff,
    sections: letter.sections.map((s) => ({
      id: s.id,
      title: s.title,
      paragraphs: s.paragraphs.map((p) => ({ id: p.id, text: p.text })),
    })),
  });
}

/**
 * Move the paragraph at `from` to `to`, both positions as they are before the
 * move. Moving down within one section shifts the target up by one once the
 * paragraph has been taken out, which is corrected here so a drop lands where
 * the indicator showed.
 */
export function moveParagraph(letter, from, to) {
  const src = letter.sections[from.section];
  const dst = letter.sections[to.section];
  if (!src || !dst || !src.paragraphs[from.index]) return;
  let index = to.index;
  if (from.section === to.section) {
    if (index === from.index || index === from.index + 1) return;
    if (index > from.index) index -= 1;
  }
  const [para] = src.paragraphs.splice(from.index, 1);
  dst.paragraphs.splice(index, 0, para);
}

/** Put a new paragraph holding `text` at `to`. */
export function insertParagraph(letter, to, text) {
  const dst = letter.sections[to.section];
  if (!dst) return;
  dst.paragraphs.splice(to.index, 0, newParagraph(text));
}
