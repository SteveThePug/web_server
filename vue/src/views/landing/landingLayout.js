/**
 * The pre-canvas landing page (a bio, an about paragraph and a row of links)
 * redrawn as boxes, for the "Import old landing page" button on its canvas.
 *
 * The old page was a 576px column (max-w-xl) centred in the window, 64px
 * from the top, with 48px between sections. On the 1400px design canvas that
 * centre is x = 412; the heights are the old sections' rendered heights,
 * rounded up. Boxes are unframed, as the old page had no frames.
 */
import { linkHtml } from "@/views/home3/templates";

const COLUMN = { x: 412, w: 576 };
const GAP = 48;

// Header.vue's heading-with-a-rule and Paragraph.vue's padded <p>, restated
// for the frame, which inherits none of the site's CSS.
const SECTION_STYLE = `<style>
  body { padding: 0; font-size: 16px; }
  h1 { margin: .25rem; padding-bottom: 2px; border-bottom: 1px solid var(--color-primary); font-size: 1.5rem; }
  p { margin: 1rem 0; padding: .25rem; line-height: 1.5; }
  a { font-weight: bold; font-style: italic; text-decoration: none; }
  a:hover { color: var(--color-tertiary); }
</style>`;

const section = (heading, body) => `${SECTION_STYLE}
<h1>${heading}</h1>
<p>
  ${body}
</p>`;

const SECTIONS = [
  {
    title: "Bio",
    h: 170,
    html: section(
      "Adam French",
      `Junior software engineer focused on full-stack development, systems
  programming, and infrastructure. First Class Honours in Computer
  Science with Mathematics from Leeds and Waterloo.`,
    ),
  },
  {
    title: "About",
    h: 200,
    html: section(
      "About",
      `This website is self-hosted and has a lot more on it than it needs to.
  Please have a look at my <a href="/cv" target="_top">CV</a> for a full
  breakdown of my experience, projects, and skills. Please visit
  <a href="/stp" target="_top">STP</a> for the prefered but less
  professional experience.`,
    ),
  },
  {
    title: "Links",
    h: 110,
    html: linkHtml([
      { name: "CV", href: "/cv" },
      { name: "STP", href: "/stp" },
      { name: "Shrines", href: "/shrines" },
      { name: "Jobs", href: "/jobs/" },
      { name: "Tracker", href: "/tracker/" },
      { name: "USACO", href: "/usaco/" },
      { name: "Email", href: "mailto:adam.a.french@outlook.com" },
      { name: "GitHub", href: "https://github.com/SteveThePug" },
      { name: "Gitea", href: "/gitea/explore/repos" },
      { name: "Spotify", href: "https://open.spotify.com/user/stevethepug" },
    ]),
  },
];

export function landingLayout() {
  let y = 64;
  return SECTIONS.map(({ title, h, html }, i) => {
    const box = { kind: "html", title, html, framed: false, ...COLUMN, y, h, z: i + 1 };
    y += h + GAP;
    return box;
  });
}
