// ==UserScript==
// @name         Job tracker capture
// @namespace    https://www.adam-french.co.uk/
// @version      1.0.1
// @description  Add the job posting you are looking at to the tracker at /cv/jobs.
// @match        *://*/*
// @noframes
// @run-at       document-idle
// @grant        GM_xmlhttpRequest
// @grant        GM_getValue
// @grant        GM_setValue
// @grant        GM_registerMenuCommand
// @connect      www.adam-french.co.uk
// @connect      adam-french.co.uk
// @connect      localhost
// @downloadURL  https://www.adam-french.co.uk/job-capture.user.js
// @updateURL    https://www.adam-french.co.uk/job-capture.user.js
// ==/UserScript==

/*
 * Browser half of the job tracker's capture feature. The server half is
 * backend/handlers/handle_job_capture.go (three token-guarded REST routes) and
 * python/app/jobs.py (the Claude call that reads the page).
 *
 * INSTALL: open /cv/jobs and press "Install capture script", or visit
 * /job-capture.user.js directly — Violentmonkey offers to install any URL
 * ending in .user.js. It lives in vue/public/ so the site serves it, and the
 * @updateURL above lets Violentmonkey pick up new versions by itself. That
 * only happens when @version changes, so BUMP @version WITH EVERY EDIT.
 *
 * CONFIGURE: Violentmonkey menu -> "Configure job capture". It asks for the
 * API base (default below; http://localhost/api for the dev stack) and the
 * token, which is generated on the site's admin panel (/admin, "Job capture
 * token") and shown there once. Both are kept in Violentmonkey's own storage. THE TOKEN MUST NEVER BE WRITTEN INTO THIS FILE:
 * it is served publicly.
 *
 * WHAT RUNS WHERE: the script matches every site so that the menu command
 * "Capture this job" works on any company's careers page, but on a page that
 * is not a known job site it does nothing at all until that command is used —
 * no button, no request. On the known sites (JOB_SITES) it shows a button and
 * asks the tracker whether the posting is already there.
 *
 * Requests go through GM_xmlhttpRequest, which is exempt from CORS (the site
 * sends no CORS headers), with `anonymous` set so no cookies travel: the
 * bearer token is the only credential.
 */

(function () {
  "use strict";

  const DEFAULT_BASE = "https://www.adam-french.co.uk/api";

  // Hosts where the button appears unprompted. Add an ATS here when one turns
  // up; nothing else needs to change, because the fields are read by Claude
  // from the page text rather than by per-site selectors.
  const JOB_SITES = [
    /(^|\.)linkedin\.com$/,
    /(^|\.)greenhouse\.io$/,
    /(^|\.)lever\.co$/,
    /(^|\.)ashbyhq\.com$/,
    /(^|\.)myworkdayjobs\.com$/,
    /(^|\.)workable\.com$/,
    /(^|\.)smartrecruiters\.com$/,
    /(^|\.)teamtailor\.com$/,
    /(^|\.)bamboohr\.com$/,
    /(^|\.)recruitee\.com$/,
    /(^|\.)breezy\.hr$/,
    /(^|\.)icims\.com$/,
    /(^|\.)jobvite\.com$/,
    /(^|\.)indeed\.com$/,
    /(^|\.)welcometothejungle\.com$/,
  ];

  // Where a posting's text usually lives, most specific first. This is a hint,
  // not a parser: it only narrows what is sent so that a search page's list of
  // other jobs does not drown out the open one. If nothing matches, the whole
  // body is sent and Claude copes.
  const CONTENT_SELECTORS = [
    ".jobs-search__job-details--container", // LinkedIn, search/collections pane
    ".jobs-details",
    ".job-view-layout",
    "main",
    '[role="main"]',
    "article",
  ];

  // Mirrors maxPageRunes in handle_job_capture.go, which cuts again anyway.
  const MAX_TEXT = 40000;

  const STATUS_LABELS = {
    saved: "Saved",
    applied: "Applied",
    screening: "Screening",
    assessment: "Assessment",
    interviewing: "Interviewing",
    offer: "Offer",
    rejected: "Rejected",
    withdrawn: "Withdrawn",
  };

  function onJobSite() {
    const host = location.hostname.toLowerCase();
    if (!JOB_SITES.some((re) => re.test(host))) return false;
    // LinkedIn is mostly not job pages.
    if (/(^|\.)linkedin\.com$/.test(host))
      return location.pathname.startsWith("/jobs");
    return true;
  }

  function pageText() {
    for (const selector of CONTENT_SELECTORS) {
      const el = document.querySelector(selector);
      const text = el?.innerText?.trim() ?? "";
      // A near-empty match is a layout shell, not the posting.
      if (text.length > 200) return text.slice(0, MAX_TEXT);
    }
    return (document.body.innerText ?? "").trim().slice(0, MAX_TEXT);
  }

  // --- configuration -------------------------------------------------------

  const config = {
    get base() {
      return (GM_getValue("base", "") || DEFAULT_BASE).replace(/\/+$/, "");
    },
    get token() {
      return GM_getValue("token", "");
    },
  };

  function configure() {
    const base = prompt(
      "Job tracker API base (no trailing slash)",
      config.base,
    );
    if (base === null) return false;
    const token = prompt(
      "Capture token (generate one on the site's admin panel)",
      config.token,
    );
    if (token === null) return false;
    GM_setValue("base", base.trim());
    GM_setValue("token", token.trim());
    return true;
  }

  // --- API -----------------------------------------------------------------

  function api(method, path, body) {
    return new Promise((resolve, reject) => {
      GM_xmlhttpRequest({
        method,
        url: config.base + path,
        headers: {
          Authorization: `Bearer ${config.token}`,
          "Content-Type": "application/json",
        },
        data: body ? JSON.stringify(body) : undefined,
        anonymous: true,
        // Above the server's own two-minute wait on Claude, so its error
        // arrives rather than this one.
        timeout: 130000,
        onload(res) {
          let json = null;
          try {
            json = JSON.parse(res.responseText);
          } catch {
            // Not JSON: an nginx error page. The status says enough.
          }
          if (res.status >= 200 && res.status < 300) return resolve(json);
          const hint =
            res.status === 401
              ? "token rejected — use Configure job capture"
              : (json?.error ?? `HTTP ${res.status}`);
          reject(new Error(hint));
        },
        onerror: () => reject(new Error(`cannot reach ${config.base}`)),
        ontimeout: () => reject(new Error("timed out")),
      });
    });
  }

  // --- UI ------------------------------------------------------------------
  // Everything lives in a closed shadow root so the host page's CSS cannot
  // restyle it and its scripts cannot read the fields. Nodes are built with
  // createElement throughout: some of these sites enforce Trusted Types, under
  // which innerHTML throws.

  const CSS = `
    :host { all: initial; }
    * { box-sizing: border-box; font-family: system-ui, sans-serif; }
    .fab {
      position: fixed; right: 16px; bottom: 16px; z-index: 2147483647;
      padding: 8px 14px; border: 1px solid #1f2937; border-radius: 999px;
      background: #fdfcf8; color: #1f2937; font-size: 13px; font-weight: 600;
      cursor: pointer; box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
    }
    .fab.tracked { background: #1f2937; color: #fdfcf8; }
    .panel {
      position: fixed; right: 16px; bottom: 60px; z-index: 2147483647;
      width: 340px; max-width: calc(100vw - 32px);
      max-height: calc(100vh - 80px); overflow: auto;
      padding: 12px; border: 1px solid #1f2937; border-radius: 8px;
      background: #fdfcf8; color: #1f2937; font-size: 13px;
      box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
      display: flex; flex-direction: column; gap: 8px;
    }
    .panel[hidden] { display: none; }
    h2 { margin: 0; font-size: 14px; }
    label { display: flex; flex-direction: column; gap: 2px; font-size: 11px; color: #6b7280; }
    input, textarea, select {
      padding: 5px 7px; border: 1px solid #d1d5db; border-radius: 4px;
      background: #fff; color: #111827; font-size: 13px; width: 100%;
    }
    textarea { resize: vertical; min-height: 44px; }
    .description { min-height: 120px; font-size: 11px; }
    summary { cursor: pointer; font-size: 11px; color: #6b7280; }
    .row { display: flex; gap: 8px; align-items: center; }
    .row > button { flex: 1; }
    button.action {
      padding: 6px 10px; border: 1px solid #1f2937; border-radius: 4px;
      background: #fff; color: #1f2937; font-size: 13px; cursor: pointer;
    }
    button.primary { background: #1f2937; color: #fdfcf8; }
    button:disabled { opacity: 0.5; cursor: default; }
    .message { min-height: 16px; font-size: 12px; color: #6b7280; }
    .message.error { color: #b91c1c; }
  `;

  function h(tag, props = {}, ...children) {
    const el = document.createElement(tag);
    for (const [key, value] of Object.entries(props)) {
      if (key === "class") el.className = value;
      else if (key.startsWith("on")) el.addEventListener(key.slice(2), value);
      else el[key] = value;
    }
    el.append(...children);
    return el;
  }

  let ui = null;

  function buildUi() {
    if (ui) return ui;

    const host = document.createElement("div");
    const root = host.attachShadow({ mode: "closed" });

    // A constructed stylesheet is not subject to the page's style-src CSP,
    // which would otherwise drop an injected <style> on the stricter sites.
    // It is not available from every userscript sandbox, hence the fallback.
    try {
      const sheet = new CSSStyleSheet();
      sheet.replaceSync(CSS);
      root.adoptedStyleSheets = [sheet];
    } catch {
      root.append(h("style", { textContent: CSS }));
    }

    const field = (label, control) => h("label", {}, label, control);
    const f = {
      jobTitle: h("input", { type: "text" }),
      company: h("input", { type: "text" }),
      location: h("input", { type: "text" }),
      url: h("input", { type: "text" }),
      status: h(
        "select",
        {},
        h("option", { value: "applied", textContent: "Applied" }),
        h("option", { value: "saved", textContent: "Saved for later" }),
      ),
      notes: h("textarea", {}),
      description: h("textarea", { class: "description" }),
    };
    const message = h("div", { class: "message" });
    const save = h("button", {
      class: "action primary",
      textContent: "Save",
      onclick: submit,
    });
    const panel = h(
      "div",
      { class: "panel", hidden: true },
      h("h2", { textContent: "Add to job tracker" }),
      field("Job title", f.jobTitle),
      field("Company", f.company),
      field("Location", f.location),
      field("Status", f.status),
      field("Notes", f.notes),
      field("URL", f.url),
      h(
        "details",
        {},
        h("summary", { textContent: "Saved description (page text)" }),
        f.description,
      ),
      message,
      h(
        "div",
        { class: "row" },
        save,
        h("button", {
          class: "action",
          textContent: "Close",
          onclick: () => (panel.hidden = true),
        }),
      ),
    );
    const fab = h("button", {
      class: "fab",
      textContent: "+ Track job",
      onclick: () => (panel.hidden ? openPanel() : (panel.hidden = true)),
    });
    root.append(fab, panel);

    // Typing in the popup must not reach the page: several of these sites
    // bind single-key shortcuts on document.
    for (const type of ["keydown", "keyup", "keypress"]) {
      host.addEventListener(type, (e) => {
        if (type === "keydown" && e.key === "Escape") panel.hidden = true;
        e.stopPropagation();
      });
    }

    document.documentElement.append(host);
    ui = { host, fab, panel, f, message, save };
    return ui;
  }

  // Inline style, not the `hidden` attribute: `:host { all: initial }` resets
  // display, which outranks the browser's own [hidden] rule.
  function setVisible(visible) {
    ui.host.style.display = visible ? "" : "none";
  }

  function say(text, isError = false) {
    ui.message.textContent = text;
    ui.message.className = isError ? "message error" : "message";
  }

  /** Reflect a tracked application on the button; null puts it back to "add". */
  function showTracked(application) {
    if (!ui) return;
    ui.fab.classList.toggle("tracked", Boolean(application));
    ui.fab.textContent = application
      ? `✓ ${STATUS_LABELS[application.status] ?? application.status}`
      : "+ Track job";
  }

  // Each open bumps this, so a slow extract for a posting that has since been
  // navigated away from cannot fill the popup for the next one.
  let openSeq = 0;

  async function openPanel() {
    if (!config.token && !configure()) return;
    buildUi();
    setVisible(true);
    const seq = ++openSeq;
    const { f, panel, save } = ui;

    for (const key of ["jobTitle", "company", "location", "notes"])
      f[key].value = "";
    f.url.value = location.href;
    f.description.value = pageText();
    f.status.value = "applied";
    save.disabled = false;
    panel.hidden = false;
    f.jobTitle.focus();

    say("Reading the page…");
    try {
      const posting = await api("POST", "/jobs/capture/extract", {
        url: location.href,
        title: document.title,
        text: f.description.value,
      });
      if (seq !== openSeq) return;
      // Only empty fields are filled: the popup is usable while Claude is
      // thinking, and anything typed in the meantime wins.
      for (const key of ["jobTitle", "company", "location", "notes"])
        if (!f[key].value.trim()) f[key].value = posting?.[key] ?? "";
      say(
        posting?.jobTitle && posting?.company
          ? "Check the fields, then save."
          : "Could not find a posting here — fill it in by hand.",
      );
    } catch (err) {
      if (seq !== openSeq) return;
      say(`Could not read the page (${err.message}). Fill it in by hand.`, true);
    }
  }

  async function submit() {
    const { f, save, panel } = ui;
    if (!f.jobTitle.value.trim() || !f.company.value.trim()) {
      say("Job title and company are required.", true);
      return;
    }
    save.disabled = true;
    say("Saving…");
    try {
      const result = await api("POST", "/jobs/capture", {
        jobTitle: f.jobTitle.value,
        company: f.company.value,
        location: f.location.value,
        url: f.url.value,
        status: f.status.value,
        notes: f.notes.value,
        description: f.description.value,
      });
      showTracked(result.application);
      say(
        {
          created: "Added to the tracker.",
          updated: "Already tracked — updated.",
          unchanged: "Already tracked — nothing to change.",
        }[result.action] ?? "Saved.",
      );
      // Invalidate any extract still in flight, then get out of the way.
      const seq = ++openSeq;
      setTimeout(() => {
        if (seq === openSeq) panel.hidden = true;
      }, 1500);
    } catch (err) {
      say(`Not saved: ${err.message}`, true);
    } finally {
      save.disabled = false;
    }
  }

  /** Ask the tracker whether the posting at the current URL is already in it. */
  async function refreshStatus() {
    showTracked(null);
    if (!config.token) return;
    const url = location.href;
    try {
      const res = await api(
        "GET",
        `/jobs/capture/status?url=${encodeURIComponent(url)}`,
      );
      // The answer is for the URL that was asked about, which on a
      // single-page site may no longer be the one on screen.
      if (url === location.href) showTracked(res?.found ? res.application : null);
    } catch {
      // The badge is a convenience; a failed lookup just leaves "+ Track job".
    }
  }

  // --- wiring --------------------------------------------------------------

  GM_registerMenuCommand("Capture this job", openPanel);
  GM_registerMenuCommand("Configure job capture", () => {
    if (configure() && ui) refreshStatus();
  });

  if (onJobSite()) {
    buildUi();
    refreshStatus();
  }

  // LinkedIn and most ATS boards are single-page apps: moving between postings
  // changes the URL without loading a page, and there is no event for that
  // which a userscript can rely on across sandboxes. Polling the address is
  // crude but dependable, and costs one string comparison a second.
  let lastHref = location.href;
  setInterval(() => {
    if (location.href === lastHref) return;
    lastHref = location.href;
    if (onJobSite()) {
      buildUi();
      setVisible(true);
      refreshStatus();
    } else if (ui && ui.panel.hidden) {
      // Wandered off the jobs section (LinkedIn feed, say): drop the button,
      // unless the popup is open with unsaved typing in it.
      setVisible(false);
    }
  }, 1000);
})();
