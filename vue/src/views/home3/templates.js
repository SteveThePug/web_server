/**
 * Starter HTML for the "New box" menu. Each runs inside the sandboxed iframe
 * (see HtmlBox.vue), so scripts work but can only reach APIs that allow
 * cross-origin requests from an opaque origin — i.e. that answer with
 * `Access-Control-Allow-Origin: *` and need no key. Open-Meteo is one.
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

/** Menu entries; `size` is [w, h] in canvas px. */
export const TEMPLATES = [
  { label: "Clock", title: "Clock", html: CLOCK, size: [240, 120] },
  { label: "Weather", title: "Weather", html: WEATHER, size: [240, 170] },
  { label: "Image / GIF", title: "Image", html: imageHtml("/img/memes/fire-woman.gif"), size: [240, 240], framed: false },
  { label: "Empty HTML", title: "HTML", html: EMPTY, size: [300, 200] },
];

export { imageHtml };
