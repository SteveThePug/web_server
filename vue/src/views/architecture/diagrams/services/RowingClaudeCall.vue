<!--
  The rowing photo reader end to end, as a sequence: the admin's upload, the Go
  handler's checks, the direct Go→Python call, how claude.py shapes the request
  for the configured model, and the sanity bounds before the insert.
  Reflects: vue/src/views/admin/CreateRowing.vue, backend/main.go (admin group),
  backend/handlers/handle_rowing.go, backend/services/python_call.go,
  python/app/main.py (rowing_read, _internal_only, _claude_failed),
  python/app/rowing.py, python/app/claude.py (MODEL, _model_options, ask_json).
-->
<template>
  <DiagramFigure id="services-rowing-claude-call" title="Reading a rowing display: browser to Claude and back" :min-width="790">
    <svg class="dg" viewBox="0 0 1120 770" role="img" aria-labelledby="services-rowing-claude-call-t">
      <title id="services-rowing-claude-call-t">Sequence of the admin rowing upload through nginx, the Go backend, Postgres, the Python service and the Claude API</title>

      <!-- Lifeline heads -->
      <rect class="dg-box dg-box--pink" x="15" y="20" width="130" height="50" rx="4" />
      <text class="dg-title dg-title--pink" x="80" y="42" text-anchor="middle">BROWSER</text>
      <text class="dg-code" x="80" y="60" text-anchor="middle">CreateRowing.vue</text>
      <rect class="dg-box" x="185" y="20" width="130" height="50" rx="4" />
      <text class="dg-title" x="250" y="42" text-anchor="middle">NGINX</text>
      <text class="dg-code" x="250" y="60" text-anchor="middle">/api/ location</text>
      <rect class="dg-box dg-box--green" x="355" y="20" width="150" height="50" rx="4" />
      <text class="dg-title" x="430" y="42" text-anchor="middle">GO BACKEND</text>
      <text class="dg-code" x="430" y="60" text-anchor="middle">CreateRowing</text>
      <rect class="dg-box" x="535" y="20" width="130" height="50" rx="4" />
      <text class="dg-title" x="600" y="42" text-anchor="middle">POSTGRES</text>
      <text class="dg-code" x="600" y="60" text-anchor="middle">models.Rowing</text>
      <rect class="dg-box dg-box--teal" x="720" y="20" width="140" height="50" rx="4" />
      <text class="dg-title" x="790" y="42" text-anchor="middle">PYTHON</text>
      <text class="dg-code" x="790" y="60" text-anchor="middle">rowing_read</text>
      <rect class="dg-box dg-box--muted" x="935" y="20" width="150" height="50" rx="4" />
      <text class="dg-title" x="1010" y="42" text-anchor="middle">CLAUDE API</text>
      <text class="dg-code" x="1010" y="60" text-anchor="middle">messages.create</text>

      <line class="dg-lifeline" x1="80" y1="70" x2="80" y2="750" />
      <line class="dg-lifeline" x1="250" y1="70" x2="250" y2="750" />
      <line class="dg-lifeline" x1="430" y1="70" x2="430" y2="750" />
      <line class="dg-lifeline" x1="600" y1="70" x2="600" y2="750" />
      <line class="dg-lifeline" x1="790" y1="70" x2="790" y2="750" />
      <line class="dg-lifeline" x1="1010" y1="70" x2="1010" y2="750" />

      <rect class="dg-activation" x="425" y="150" width="10" height="580" />
      <rect class="dg-activation" x="785" y="330" width="10" height="282" />
      <rect class="dg-activation" x="1005" y="458" width="10" height="40" />

      <!-- 1 upload -->
      <text class="dg-edge-label dg-edge-label--pink" x="165" y="94" text-anchor="middle">POST /api/rowing</text>
      <text class="dg-edge-label" x="165" y="107" text-anchor="middle">multipart field "image"</text>
      <path class="dg-edge dg-edge--pink" d="M80,116 H246" />
      <text class="dg-edge-label" x="340" y="128" text-anchor="middle">POST /rowing</text>
      <text class="dg-edge-label" x="340" y="141" text-anchor="middle">/api stripped</text>
      <path class="dg-edge" d="M250,150 H421" />

      <!-- 2 checks -->
      <rect class="dg-box dg-box--tint" x="445" y="160" width="270" height="74" rx="3" />
      <text class="dg-small" x="455" y="180">AuthMiddlewear → AdminMiddleware</text>
      <text class="dg-small" x="455" y="198">exif.Decode → DateTime, else 400</text>
      <text class="dg-small" x="455" y="216">≤ 10MB · jpeg / png / gif / webp only</text>

      <!-- 3 duplicate -->
      <text class="dg-edge-label" x="515" y="256" text-anchor="middle">WHERE date = ?</text>
      <path class="dg-edge" d="M435,264 H596" />
      <text class="dg-edge-label" x="515" y="284" text-anchor="middle">not found (found → 409)</text>
      <path class="dg-edge dg-edge--muted dg-edge--dashed" d="M600,292 H439" />

      <!-- 4 call python -->
      <text class="dg-edge-label dg-edge-label--mint" x="610" y="312" text-anchor="middle">services.CallPython · POST /internal/rowing/read</text>
      <text class="dg-edge-label" x="610" y="325" text-anchor="middle">{media_type, data: base64} · client timeout 2 min</text>
      <path class="dg-edge dg-edge--green dg-edge--flow" d="M435,334 H781" />

      <!-- 5 shape request -->
      <rect class="dg-box dg-box--pink dg-box--tint" x="805" y="344" width="300" height="94" rx="3" />
      <text class="dg-text dg-text--pink" x="815" y="362">claude._model_options(MODEL, 256)</text>
      <text class="dg-small" x="815" y="380">MODEL = CLAUDE_MODEL or claude-sonnet-5-5</text>
      <text class="dg-small" x="815" y="398">claude-haiku-4-5 → max_tokens 256</text>
      <text class="dg-small" x="815" y="416">any other → max_tokens 256 + 4000,</text>
      <text class="dg-small" x="815" y="432">output_config.effort "low"</text>

      <!-- 6 create -->
      <text class="dg-edge-label" x="900" y="455" text-anchor="middle">image block + PROMPT</text>
      <text class="dg-edge-label" x="900" y="468" text-anchor="middle">30s timeout · 2 SDK retries</text>
      <path class="dg-edge" d="M795,474 H1001" />
      <text class="dg-edge-label" x="900" y="496" text-anchor="middle">reply: content blocks</text>
      <path class="dg-edge dg-edge--muted dg-edge--dashed" d="M1005,504 H799" />

      <!-- 7 parse -->
      <rect class="dg-box dg-box--tint" x="805" y="516" width="300" height="62" rx="3" />
      <text class="dg-small" x="815" y="534">first text block → strip_markdown_fence</text>
      <text class="dg-small" x="815" y="550">→ RowingReading (missing keys = 0)</text>
      <text class="dg-small" x="815" y="566">refusal / no text / bad JSON → ReadError</text>

      <text class="dg-edge-label dg-edge-label--mint" x="610" y="590" text-anchor="middle">200 {timeMinutes, timeSeconds, distance}</text>
      <text class="dg-edge-label dg-edge-label--pink" x="610" y="603" text-anchor="middle">or 502 {detail} → *PythonError</text>
      <path class="dg-edge dg-edge--muted dg-edge--dashed" d="M785,612 H439" />

      <!-- 8 bounds -->
      <rect class="dg-box dg-box--tint" x="445" y="622" width="270" height="62" rx="3" />
      <text class="dg-small" x="455" y="640">distance 100–100000 m, time 30–7200 s</text>
      <text class="dg-small" x="455" y="656">pace 80–150 s per 500 m, else 400</text>
      <text class="dg-small" x="455" y="672">calories = distance / 7500 × 500</text>

      <!-- 9 insert + reply -->
      <text class="dg-edge-label" x="515" y="700" text-anchor="middle">INSERT</text>
      <path class="dg-edge" d="M435,706 H596" />
      <text class="dg-edge-label dg-edge-label--pink" x="255" y="716" text-anchor="middle">201 rowing · or 500 {error: Python's detail}</text>
      <path class="dg-edge dg-edge--pink dg-edge--dashed" d="M425,726 H84" />

      <!-- badges -->
      <circle class="dg-badge" cx="22" cy="116" r="9" /><text class="dg-badge-text" x="22" y="116">1</text>
      <circle class="dg-badge" cx="22" cy="190" r="9" /><text class="dg-badge-text" x="22" y="190">2</text>
      <circle class="dg-badge" cx="22" cy="264" r="9" /><text class="dg-badge-text" x="22" y="264">3</text>
      <circle class="dg-badge" cx="22" cy="334" r="9" /><text class="dg-badge-text" x="22" y="334">4</text>
      <circle class="dg-badge" cx="22" cy="394" r="9" /><text class="dg-badge-text" x="22" y="394">5</text>
      <circle class="dg-badge" cx="22" cy="474" r="9" /><text class="dg-badge-text" x="22" y="474">6</text>
      <circle class="dg-badge" cx="22" cy="548" r="9" /><text class="dg-badge-text" x="22" y="548">7</text>
      <circle class="dg-badge" cx="22" cy="652" r="9" /><text class="dg-badge-text" x="22" y="652">8</text>
      <circle class="dg-badge" cx="22" cy="716" r="9" /><text class="dg-badge-text" x="22" y="716">9</text>
    </svg>
    <template #caption>
      Everything that needs the session or the database stays in Go (steps 2, 3, 8, 9); Python
      only turns pixels into three numbers. The duplicate check (3) runs <em>before</em> the paid call,
      keyed on the photo's EXIF timestamp, so a re-upload costs nothing. Step 5 is where
      <code>CLAUDE_MODEL</code> matters: models that think get 4,000 tokens of headroom and
      <code>effort: "low"</code>, while Haiku 4.5 rejects the effort setting and gets neither. The
      timeouts nest on purpose: Python's 30s per attempt with two SDK retries sits inside Go's
      2-minute client, so a slow Claude surfaces as Python's fixed message, which Go relays to the
      admin only when it came back as a 502.
    </template>
  </DiagramFigure>
</template>

<script setup>
import DiagramFigure from "../../DiagramFigure.vue";
</script>
