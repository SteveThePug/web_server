/**
 * The built-in starter HTML. The "+ New box" menu no longer reads TEMPLATES
 * directly: it lists saved components, and TEMPLATES is what the Components
 * panel's "Import built-ins" turns into components (stores/components.js), so
 * they can be edited and deleted like any other. Changing an entry here only
 * reaches a site whose admin imports it again (after deleting the old one).
 * The helpers at the bottom are still used directly, by dropped uploads and
 * the starter layouts.
 *
 * Each runs inside the sandboxed iframe (see HtmlBox.vue), so scripts work but
 * can only reach APIs that allow cross-origin requests from an opaque origin —
 * i.e. that answer with `Access-Control-Allow-Origin: *` and need no key.
 * Open-Meteo is one.
 *
 * The site's colour and font tokens are already defined in the frame
 * (frameStyles.js), so templates use var(--color-…) freely.
 */

const CLOCK = `<style>
  body { display: grid; place-items: center; text-align: center; }
  .time { font-family: var(--font-heading); color: var(--color-primary); font-size: 3rem; line-height: 1; }
  .date { color: var(--color-tertiary); }
</style>
<div>
  <div class="time" id="time"></div>
  <div class="date" id="date"></div>
</div>
<script>
  const zone = "Europe/London";
  const time = new Intl.DateTimeFormat("en-GB", { timeZone: zone, hour: "2-digit", minute: "2-digit", second: "2-digit" });
  const date = new Intl.DateTimeFormat("en-GB", { timeZone: zone, weekday: "long", day: "numeric", month: "long" });
  function tick() {
    const now = new Date();
    document.getElementById("time").textContent = time.format(now);
    document.getElementById("date").textContent = date.format(now);
  }
  tick();
  setInterval(tick, 1000);
</script>`;

const WEATHER = `<style>
  body { display: grid; place-items: center; text-align: center; }
  .temp { font-family: var(--font-heading); color: var(--color-primary); font-size: 3rem; line-height: 1; }
  .desc { color: var(--color-secondary); }
  .meta { color: var(--color-tertiary); font-size: .8em; }
</style>
<div>
  <h3>London</h3>
  <div class="temp" id="temp">…</div>
  <div class="desc" id="desc"></div>
  <div class="meta" id="meta"></div>
</div>
<script>
  // https://open-meteo.com/en/docs — free, keyless, CORS-open.
  const URL_ = "https://api.open-meteo.com/v1/forecast"
    + "?latitude=51.5072&longitude=-0.1276"
    + "&current=temperature_2m,apparent_temperature,weather_code,wind_speed_10m"
    + "&wind_speed_unit=mph&timezone=Europe%2FLondon";

  // WMO weather interpretation codes, as Open-Meteo documents them.
  const CODES = {
    0: "Clear sky", 1: "Mainly clear", 2: "Partly cloudy", 3: "Overcast",
    45: "Fog", 48: "Rime fog", 51: "Light drizzle", 53: "Drizzle", 55: "Heavy drizzle",
    56: "Freezing drizzle", 57: "Freezing drizzle", 61: "Light rain", 63: "Rain",
    65: "Heavy rain", 66: "Freezing rain", 67: "Freezing rain", 71: "Light snow",
    73: "Snow", 75: "Heavy snow", 77: "Snow grains", 80: "Light showers",
    81: "Showers", 82: "Violent showers", 85: "Snow showers", 86: "Snow showers",
    95: "Thunderstorm", 96: "Thunderstorm, hail", 99: "Thunderstorm, hail",
  };

  async function load() {
    try {
      const res = await fetch(URL_);
      const { current } = await res.json();
      document.getElementById("temp").textContent = Math.round(current.temperature_2m) + "°C";
      document.getElementById("desc").textContent = CODES[current.weather_code] ?? "";
      document.getElementById("meta").textContent =
        "feels " + Math.round(current.apparent_temperature) + "° · wind " + Math.round(current.wind_speed_10m) + " mph";
    } catch (err) {
      document.getElementById("desc").textContent = "Weather unavailable";
    }
  }
  load();
  setInterval(load, 10 * 60 * 1000);
</script>`;

const imageHtml = (src) => `<style>
  body { padding: 0; overflow: hidden; }
  img { display: block; width: 100%; height: 100%; object-fit: cover; }
</style>
<img src="${src}" alt="">`;

const EMPTY = `<h3>New box</h3>
<p>Edit me.</p>`;

// A sticky note for reminders to self. Unframed, since the note is its own
// frame; the menu entry makes it adminOnly.
const NOTE = `<style>
  body { padding: 0; }
  .note {
    box-sizing: border-box; height: 100%; padding: .75rem 1rem;
    background: #f6e27a; color: #2b2610; transform: rotate(-1deg);
    box-shadow: 2px 3px 0 rgba(0, 0, 0, .35); overflow: auto;
  }
  .note h3 { margin: 0 0 .3rem; color: inherit; font-family: var(--font-heading); }
  .note ul { margin: 0; padding-left: 1.1rem; }
</style>
<div class="note">
  <h3>TODO</h3>
  <ul>
    <li>Something to implement</li>
  </ul>
</div>`;

const HEADING = `<style>
  body { display: grid; place-items: center; text-align: center; }
  h1 { margin: 0; font-family: var(--font-heading); color: var(--color-primary); font-size: 3.5rem; line-height: 1; }
  p { margin: .25rem 0 0; color: var(--color-tertiary); }
</style>
<div>
  <h1>Title</h1>
  <p>subtitle</p>
</div>`;

const TEXT = `<h3>Heading</h3>
<p>
  A paragraph about it. <b>Bold</b>, <i>italic</i> and
  <a href="https://example.com" target="_blank">links</a> all work.
</p>
<blockquote>“A quote you like.”</blockquote>`;

// Links. An HTML box is its own document in a sandboxed frame, so a link
// must name where it opens: target="_top" replaces the site page (a full
// load, not a router navigation — fine for leaving the page), and
// target="_blank" a new tab. Without a target it would load inside the box.
const LINK_STYLE = `<style>
  body { display: grid; place-items: center; padding: 4px; }
  .links { display: flex; flex-wrap: wrap; justify-content: space-around; gap: 1rem; }
  a {
    padding: .5rem 1rem; border: 1px solid currentColor;
    color: var(--color-primary); text-decoration: none; font-size: 16px;
    transition: color .15s ease;
  }
  a:hover { color: var(--color-tertiary); }
</style>`;

const linkHtml = (links) => `${LINK_STYLE}
<nav class="links">
${links.map(({ name, href }) => `  <a href="${href}" target="${href.startsWith("/") ? "_top" : "_blank"}">${name}</a>`).join("\n")}
</nav>`;

// A scrolling grid of images. "Insert upload" in the editor adds more.
const galleryHtml = (srcs) => `<style>
  body { padding: 4px; }
  .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 4px; }
  .grid img { width: 100%; aspect-ratio: 1; object-fit: cover; display: block; }
</style>
<div class="grid">
${srcs.map((src) => `  <img src="${src}" alt="">`).join("\n")}
</div>`;

const videoHtml = (src) => `<style>
  body { padding: 0; overflow: hidden; background: #000; }
  video { display: block; width: 100%; height: 100%; object-fit: contain; }
</style>
<video src="${src}" controls loop muted></video>`;

/**
 * Built-in starters; `label` becomes the component's name and `size` ([w, h]
 * in canvas px) its default size. `adminOnly` sets a new box's flag, which can
 * be changed afterwards in the editor.
 */
export const TEMPLATES = [
  { label: "Note (admin only)", title: "Note", html: NOTE, size: [260, 200], framed: false, adminOnly: true },
  { label: "Heading", title: "Heading", html: HEADING, size: [600, 140], framed: false },
  { label: "Text", title: "Text", html: TEXT, size: [400, 260] },
  { label: "Link", title: "Link", html: linkHtml([{ name: "CV", href: "/cv" }]), size: [120, 60], framed: false },
  {
    label: "Link row",
    title: "Links",
    html: linkHtml([
      { name: "CV", href: "/cv" },
      { name: "GitHub", href: "https://github.com/SteveThePug" },
    ]),
    size: [400, 70],
    framed: false,
  },
  { label: "Image / GIF", title: "Image", html: imageHtml("/img/memes/fire-woman.gif"), size: [240, 240], framed: false },
  { label: "Image gallery", title: "Gallery", html: galleryHtml(["/img/memes/fire-woman.gif", "/img/memes/epic.jpeg"]), size: [460, 320] },
  { label: "Video", title: "Video", html: videoHtml("/img/demoman/1760582395316219.webm"), size: [400, 260], framed: false },
  { label: "Clock", title: "Clock", html: CLOCK, size: [240, 120] },
  { label: "Weather", title: "Weather", html: WEATHER, size: [240, 170] },
  { label: "Empty HTML", title: "HTML", html: EMPTY, size: [300, 200] },
];

export { imageHtml, galleryHtml, videoHtml, linkHtml };
