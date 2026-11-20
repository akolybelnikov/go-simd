const byID = (id) => document.getElementById(id);
const storageKey = "simdlab-session";
let token = location.hash.slice(1);
if (token) {
  sessionStorage.setItem(storageKey, token);
  history.replaceState(null, "", location.pathname);
} else {
  token = sessionStorage.getItem(storageKey) || "";
}

const integer = new Intl.NumberFormat(undefined, { maximumFractionDigits: 0 });
const number = new Intl.NumberFormat(undefined, { maximumSignificantDigits: 5 });
const axisNumber = new Intl.NumberFormat(undefined, { maximumSignificantDigits: 3 });

let planRef = "";
let activeRun = null;
let busy = false;
let currentReport = "";
let currentSeries = [];

function status(message, error = false, id = "status") {
  byID(id).textContent = message;
  byID(id).dataset.error = String(error);
}

function updateControls() {
  byID("run").disabled = busy || !planRef;
  byID("starter").disabled = busy || !token;
  byID("cancel").hidden = !activeRun;
  byID("capture").disabled = busy || !token;
  byID("question").disabled = busy || !token;
  byID("ask-submit").disabled = busy || !token;
  document.querySelector(".file-label").setAttribute("aria-disabled", String(byID("capture").disabled));
  byID("results").setAttribute("aria-busy", String(busy));
}

async function api(path, body, signal) {
  const response = await fetch(path, {
    method: "POST",
    headers: { "X-Simdlab-Session": token, "Content-Type": "application/json" },
    body,
    signal,
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || "The local lab could not complete the request.");
  }
  return data;
}

function fillList(node, items) {
  node.replaceChildren(...items.map((item) => {
    const line = document.createElement("li");
    line.textContent = item;
    return line;
  }));
  node.hidden = items.length === 0;
}

function setPlan(plan) {
  planRef = plan.ref;
  byID("operation").textContent = plan.operation;
  const methods = plan.methods.length === 2 ? "Two methods" : plan.methods.length + " methods";
  byID("plan-settings").textContent = methods + " · " + integer.format(plan.size) +
    " generated values per array · " + plan.repeats + " repeats · running on this machine";
  fillList(byID("methods"), plan.methods);
  fillList(byID("plan-notes"), plan.warnings || []);
  byID("plan-host").textContent = plan.host;
  byID("model-mode").textContent = plan.ollama ?
    "Local Ollama helps with additional wording." : "Example requests work offline.";
}

async function useStarter() {
  busy = true;
  updateControls();
  status("Preparing addition…");
  try {
    setPlan(await api("/api/starter"));
    status("");
  } catch (error) {
    status(error.message, true);
  } finally {
    busy = false;
    updateControls();
  }
}

async function runComparison() {
  const ref = planRef;
  planRef = "";
  activeRun = new AbortController();
  busy = true;
  updateControls();
  status("Measuring on this machine…");
  try {
    showResult(await api("/api/run", JSON.stringify({ ref }), activeRun.signal));
    status("Comparison saved. Use the addition starter or preview another comparison to run again.");
  } catch (error) {
    if (error.name === "AbortError") {
      status("Run canceled. Open a fresh preview to run again.");
    } else {
      status(error.message, true);
    }
  } finally {
    activeRun = null;
    busy = false;
    updateControls();
  }
}

async function openCapture(event) {
  const file = event.target.files[0];
  event.target.value = "";
  if (!file) return;
  if (file.size > 8 * 1024 * 1024) {
    status("Choose a measurement capture up to 8 MiB.", true, "open-status");
    return;
  }
  busy = true;
  updateControls();
  status("Reading the capture…", false, "open-status");
  try {
    showResult(await api("/api/open", file));
    status("");
    status("Opened " + file.name, false, "open-status");
  } catch (error) {
    status(error.message, true, "open-status");
  } finally {
    busy = false;
    updateControls();
  }
}

async function askQuestion(event) {
  event.preventDefault();
  const question = byID("question").value.trim();
  if (!question || busy) return;
  busy = true;
  updateControls();
  status("Resolving the question…", false, "ask-status");
  try {
    setPlan(await api("/api/ask", JSON.stringify({ question })));
    status("");
    status("Review the new preview, then choose Run locally.", false, "ask-status");
  } catch (error) {
    status(error.message, true, "ask-status");
  } finally {
    busy = false;
    updateControls();
  }
}

function showResult(result) {
  const [headline, ...detail] = result.timing.trim().split("\n");
  byID("headline").textContent = headline;
  byID("medians").textContent = detail.join("\n");
  fillList(byID("caveats"), result.caveats || []);
  const host = result.host;
  byID("host").textContent = [host.os + " / " + host.arch, host.cpu === "unknown" ? "" : host.cpu, host.go_version,
    host.vector_bits ? host.vector_bits + "-bit vectors" : "",
    host.emulated ? "SIMD emulated" : ""].filter(Boolean).join(" · ");
  currentSeries = result.series;
  const sizes = [...new Set(currentSeries.map((series) => series.size))];
  byID("plot-size").replaceChildren(...sizes.map((size) => {
    const option = document.createElement("option");
    option.value = size;
    option.textContent = integer.format(size) + " values";
    return option;
  }));
  byID("plot-size-label").hidden = sizes.length < 2;
  byID("artifact").textContent = result.report_path ? "Saved locally at " + result.report_path : "";
  currentReport = result.report;
  byID("report").disabled = !currentReport;
  byID("empty").hidden = true;
  byID("result-body").hidden = false;
  showRepeats();
  byID("headline").focus();
}

function selectedSeries() {
  const size = Number(byID("plot-size").value);
  return currentSeries.filter((series) => series.size === size);
}

function showRepeats() {
  const rows = document.createDocumentFragment();
  for (const series of selectedSeries()) {
    for (const repeat of series.repeats) {
      const tr = document.createElement("tr");
      for (const value of [series.method, integer.format(series.size), repeat.number, number.format(repeat.ns)]) {
        const td = document.createElement("td");
        td.textContent = value;
        tr.append(td);
      }
      rows.append(tr);
    }
  }
  byID("repeat-table").replaceChildren(rows);
  drawRepeats();
}

function drawRepeats() {
  const series = selectedSeries();
  const canvas = byID("plot");
  canvas.hidden = series.length === 0 || series.length > 12;
  const width = canvas.clientWidth;
  if (canvas.hidden || width < 40) return;

  const ratio = Math.min(window.devicePixelRatio || 1, 2);
  const rowHeight = 41;
  const axis = series.length * rowHeight + 12;
  const height = axis + 26;
  canvas.width = Math.round(width * ratio);
  canvas.height = Math.round(height * ratio);
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  ctx.scale(ratio, ratio);
  ctx.font = "12px system-ui, sans-serif";
  const left = Math.min(187, Math.round(width * .42));
  const right = width - 25;
  const largest = Math.max(...series.flatMap((row) => row.repeats.map((point) => point.ns)));
  const upper = largest > 0 ? largest : 1;
  const position = (value) => left + (value / upper) * (right - left);

  ctx.strokeStyle = "#dbe3e8";
  ctx.fillStyle = "#4d626f";
  ctx.textAlign = "center";
  for (let tick = 0; tick <= 4; tick++) {
    const x = position((tick / 4) * upper);
    ctx.beginPath();
    ctx.moveTo(x, 7);
    ctx.lineTo(x, axis);
    ctx.stroke();
    ctx.fillText(axisNumber.format((tick / 4) * upper), x, axis + 16, 45);
  }
  ctx.textAlign = "left";
  series.forEach((row, index) => {
    const y = 19 + index * rowHeight;
    ctx.fillStyle = "#172b37";
    ctx.fillText(row.method, 0, y + 4, left - 8);
    ctx.fillStyle = index % 2 ? "#197468" : "#175e96";
    ctx.strokeStyle = "#ffffff";
    row.repeats.forEach((point, j) => {
      ctx.beginPath();
      ctx.arc(position(point.ns), y + ((j % 5) - 2) * 6, 4, 0, Math.PI * 2);
      ctx.fill();
      ctx.stroke();
    });
  });
}

function saveReport() {
  const url = URL.createObjectURL(new Blob([currentReport], { type: "text/markdown;charset=utf-8" }));
  const download = document.createElement("a");
  download.href = url;
  download.download = "simdlab-report.md";
  download.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

byID("starter").addEventListener("click", useStarter);
byID("run").addEventListener("click", runComparison);
byID("cancel").addEventListener("click", () => activeRun?.abort());
byID("report").addEventListener("click", saveReport);
byID("capture").addEventListener("change", openCapture);
byID("ask").addEventListener("submit", askQuestion);
byID("plot-size").addEventListener("change", showRepeats);
new ResizeObserver(drawRepeats).observe(byID("results"));

if (token) {
  useStarter();
} else {
  updateControls();
  status("Open the private browser URL printed by simdlab ui.", true);
}
