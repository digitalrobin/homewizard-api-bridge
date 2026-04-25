package main

const authPageHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>HomeWizard Bridge Pairing</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f4efe6;
      --panel: rgba(255, 252, 247, 0.92);
      --ink: #1f2a1f;
      --muted: #5f6d5f;
      --line: rgba(31, 42, 31, 0.12);
      --accent: #146356;
      --accent-strong: #0c4d42;
      --warn: #8a4b0f;
      --ok: #17633b;
      --shadow: 0 24px 70px rgba(38, 48, 38, 0.12);
    }

    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      font-family: "Avenir Next", "Segoe UI", sans-serif;
      color: var(--ink);
      background:
        radial-gradient(circle at top left, rgba(20, 99, 86, 0.15), transparent 30%),
        radial-gradient(circle at bottom right, rgba(197, 153, 88, 0.18), transparent 34%),
        linear-gradient(160deg, #f7f1e8 0%, #efe7dc 48%, #f7f4ee 100%);
      display: grid;
      place-items: center;
      padding: 24px;
    }

    .shell {
      width: min(760px, 100%);
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 28px;
      box-shadow: var(--shadow);
      overflow: hidden;
      backdrop-filter: blur(10px);
    }

    .hero {
      padding: 28px 28px 18px;
      background:
        linear-gradient(135deg, rgba(20, 99, 86, 0.12), rgba(197, 153, 88, 0.12)),
        linear-gradient(180deg, rgba(255,255,255,0.7), rgba(255,255,255,0));
      border-bottom: 1px solid var(--line);
    }

    .eyebrow {
      font-size: 12px;
      letter-spacing: 0.12em;
      text-transform: uppercase;
      color: var(--muted);
      margin-bottom: 10px;
    }

    h1 {
      margin: 0;
      font-size: clamp(28px, 5vw, 44px);
      line-height: 1.02;
    }

    .subtitle {
      margin: 12px 0 0;
      color: var(--muted);
      max-width: 52ch;
      line-height: 1.5;
    }

    .content {
      display: grid;
      gap: 18px;
      padding: 24px;
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 14px;
    }

    .card {
      border: 1px solid var(--line);
      border-radius: 20px;
      padding: 18px;
      background: rgba(255, 255, 255, 0.72);
    }

    .label {
      font-size: 12px;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      color: var(--muted);
      margin-bottom: 8px;
    }

    .value {
      font-size: 22px;
      font-weight: 600;
      line-height: 1.2;
      word-break: break-word;
    }

    .status-pill {
      display: inline-flex;
      align-items: center;
      gap: 10px;
      padding: 10px 14px;
      border-radius: 999px;
      background: rgba(20, 99, 86, 0.10);
      color: var(--accent-strong);
      font-weight: 600;
    }

    .status-pill.warn {
      background: rgba(138, 75, 15, 0.10);
      color: var(--warn);
    }

    .dot {
      width: 10px;
      height: 10px;
      border-radius: 999px;
      background: currentColor;
    }

    .actions {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
      align-items: center;
    }

    button {
      appearance: none;
      border: 0;
      border-radius: 14px;
      padding: 14px 18px;
      background: var(--accent);
      color: white;
      font: inherit;
      font-weight: 700;
      cursor: pointer;
      transition: transform 120ms ease, background 120ms ease;
    }

    button:hover { transform: translateY(-1px); background: var(--accent-strong); }
    button:disabled { opacity: 0.6; cursor: wait; transform: none; }

    .ghost {
      background: transparent;
      color: var(--accent-strong);
      border: 1px solid rgba(20, 99, 86, 0.25);
    }

    .message {
      white-space: pre-wrap;
      line-height: 1.55;
      color: var(--ink);
    }

    .message small {
      color: var(--muted);
      display: block;
      margin-top: 8px;
    }

    ol {
      margin: 0;
      padding-left: 20px;
      color: var(--ink);
      line-height: 1.55;
    }

    code {
      font-family: "SFMono-Regular", "Menlo", monospace;
      font-size: 0.95em;
      background: rgba(31, 42, 31, 0.06);
      padding: 2px 6px;
      border-radius: 8px;
    }

    @media (max-width: 640px) {
      .hero, .content { padding: 20px; }
      .actions { flex-direction: column; align-items: stretch; }
      button { width: 100%; }
    }
  </style>
</head>
<body>
  <main class="shell">
    <section class="hero">
      <div class="eyebrow">HomeWizard Bridge</div>
      <h1>Pairing Console</h1>
      <p class="subtitle">Use this page to trigger the HomeWizard v2 pairing flow, monitor token status, and confirm when the bridge is ready for Loxone.</p>
    </section>

    <section class="content">
      <div id="pairingState" class="status-pill warn">
        <span class="dot"></span>
        <span>Checking bridge status…</span>
      </div>

      <div class="grid">
        <article class="card">
          <div class="label">HomeWizard Target</div>
          <div id="target" class="value">-</div>
        </article>
        <article class="card">
          <div class="label">Bridge User</div>
          <div id="user" class="value">-</div>
        </article>
        <article class="card">
          <div class="label">Token Updated</div>
          <div id="updated" class="value">Not paired yet</div>
        </article>
      </div>

      <article class="card">
        <div class="label">Actions</div>
        <div class="actions">
          <button id="pairButton" type="button">Start / Retry Pairing</button>
          <button id="refreshButton" class="ghost" type="button">Refresh Status</button>
        </div>
      </article>

      <article class="card">
        <div class="label">Last Result</div>
        <div id="message" class="message">No pairing attempt yet.</div>
      </article>

      <article class="card">
        <div class="label">How to pair</div>
        <ol>
          <li>Press <strong>Start / Retry Pairing</strong>.</li>
          <li>If the bridge tells you pairing is not enabled yet, press the button on the HomeWizard P1 meter.</li>
          <li>Press <strong>Start / Retry Pairing</strong> again within 30 seconds.</li>
          <li>When the token is stored, the bridge is ready and Loxone can use the metric endpoints.</li>
        </ol>
      </article>
    </section>
  </main>

  <script>
    const pairingState = document.getElementById('pairingState');
    const target = document.getElementById('target');
    const user = document.getElementById('user');
    const updated = document.getElementById('updated');
    const message = document.getElementById('message');
    const pairButton = document.getElementById('pairButton');
    const refreshButton = document.getElementById('refreshButton');

    function setStatus(tokenReady) {
      pairingState.className = tokenReady ? 'status-pill' : 'status-pill warn';
      pairingState.innerHTML = tokenReady
        ? '<span class="dot"></span><span>Token ready. The bridge is paired.</span>'
        : '<span class="dot"></span><span>No token yet. Pairing still needed.</span>';
    }

    function setMessage(text, detail) {
      message.innerHTML = '';
      message.textContent = text || 'No message.';
      if (detail) {
        const small = document.createElement('small');
        small.textContent = detail;
        message.appendChild(small);
      }
    }

    async function loadStatus() {
      const response = await fetch('/auth/status', { cache: 'no-store' });
      if (!response.ok) {
        throw new Error('Could not fetch auth status.');
      }

      const data = await response.json();
      setStatus(Boolean(data.token_ready));
      target.textContent = data.homewizard || '-';
      user.textContent = data.user || '-';
      updated.textContent = data.token_updated ? new Date(data.token_updated).toLocaleString() : 'Not paired yet';
      return data;
    }

    async function doPair() {
      pairButton.disabled = true;
      setMessage('Sending pairing request…');

      try {
        const response = await fetch('/pair', { method: 'POST' });
        const data = await response.json();
        const detail = data.error || '';

        if (response.ok) {
          setMessage(data.message || 'Pairing succeeded.', detail);
        } else {
          setMessage(data.message || 'Pairing needs another step.', detail);
        }

        await loadStatus();
      } catch (error) {
        setMessage('Pairing request failed.', error.message || String(error));
      } finally {
        pairButton.disabled = false;
      }
    }

    async function refresh() {
      try {
        await loadStatus();
      } catch (error) {
        setStatus(false);
        setMessage('Status refresh failed.', error.message || String(error));
      }
    }

    pairButton.addEventListener('click', doPair);
    refreshButton.addEventListener('click', refresh);

    refresh();
    setInterval(refresh, 3000);
  </script>
</body>
</html>
`
