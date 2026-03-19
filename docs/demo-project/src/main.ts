import "./style.css";

const API_BASE = "http://localhost:3000/api/v1";

let ws: WebSocket | null = null;

function log(message: string) {
  const logs = document.querySelector("#logs")!;
  logs.innerHTML += `<div>> ${message}</div>`;
  logs.scrollTop = logs.scrollHeight;
}

document.querySelector<HTMLDivElement>("#app")!.innerHTML = `
<div class="container">
  <h1>🚀 Node Manager Demo</h1>

  <div class="card">
    <h2>1. Connect Node (WebSocket)</h2>
    <input id="clientId" placeholder="Client ID" />
    <select id="clientType">
      <option value="worker">worker</option>
      <option value="browser">browser</option>
    </select>
    <button id="connectBtn">Connect</button>
  </div>

  <div class="card">
    <h2>2. Trigger Job</h2>
    <input id="url" placeholder="https://example.com" />
    <select id="method">
      <option>GET</option>
      <option>POST</option>
    </select>
    <button id="triggerBtn">Send Job</button>
  </div>

  <div class="card">
    <h2>Logs</h2>
    <div id="logs" class="logs"></div>
  </div>
</div>
`;

// CONNECT WS
document.getElementById("connectBtn")!.addEventListener("click", () => {
  const clientId = (document.getElementById("clientId") as HTMLInputElement)
    .value;
  const clientType = (
    document.getElementById("clientType") as HTMLSelectElement
  ).value;

  if (!clientId) {
    alert("Enter clientId");
    return;
  }

  ws = new WebSocket(
    `ws://localhost:3000/api/v1/ws?clientId=${clientId}&clientType=${clientType}`,
  );

  ws.onopen = () => {
    log("✅ WebSocket connected");

    // heartbeat
    setInterval(() => {
      ws?.send(JSON.stringify({ type: "PING" }));
    }, 5000);
  };

  ws.onmessage = async (event) => {
    const msg = JSON.parse(event.data);
    log("📩 Received: " + JSON.stringify(msg));

    if (msg.type === "JOB") {
      const { jobId, payload } = msg;

      log(`⚙️ Executing job ${jobId}`);

      const start = performance.now();

      try {
        const res = await fetch(payload.url, {
          method: payload.method,
        });

        const body = await res.text();

        const result = {
          status: res.status,
          body: body.slice(0, 200), // limit
          timingMs: Math.round(performance.now() - start),
        };

        ws?.send(
          JSON.stringify({
            type: "RESULT",
            jobId,
            payload: result,
          }),
        );

        log(`✅ Job ${jobId} done`);
      } catch (err) {
        log("❌ Job failed");
      }
    }
  };

  ws.onclose = () => log("❌ WebSocket closed");
});

// TRIGGER JOB
document.getElementById("triggerBtn")!.addEventListener("click", async () => {
  const url = (document.getElementById("url") as HTMLInputElement).value;
  const method = (document.getElementById("method") as HTMLSelectElement).value;

  log("📤 Sending trigger request...");

  const res = await fetch(`${API_BASE}/trigger/single`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      link: url,
      method,
    }),
  });

  const data = await res.json();

  log("📥 API Response: " + JSON.stringify(data));
});
