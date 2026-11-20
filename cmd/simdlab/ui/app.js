const byID = (id) => document.getElementById(id);
const examples = document.querySelectorAll(".comparison-examples button");
const tabs = [...document.querySelectorAll('[role="tab"]')];
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
let currentResult;
let operations = [];
let selectedOperation;
let agentPrompt;

function selectTab(name) {
  for (const tab of tabs) {
    const selected = tab.id === "tab-" + name;
    tab.setAttribute("aria-selected", String(selected));
    tab.tabIndex = selected ? 0 : -1;
    byID(tab.getAttribute("aria-controls")).hidden = !selected;
  }
}

for (const [index, tab] of tabs.entries()) {
  tab.addEventListener("click", () => selectTab(tab.id.slice(4)));
  tab.addEventListener("keydown", (event) => {
    const target = {
      ArrowLeft: (index + tabs.length - 1) % tabs.length,
      ArrowRight: (index + 1) % tabs.length,
      Home: 0,
      End: tabs.length - 1,
    }[event.key];
    if (target === undefined) return;
    event.preventDefault();
    tabs[target].click();
    tabs[target].focus();
  });
}

function status(message, error = false, id = "status") {
  byID(id).textContent = message;
  byID(id).dataset.error = String(error);
}

function updateControls() {
  byID("run").disabled = busy || !planRef;
  byID("starter").disabled = busy || !token;
  byID("cancel").hidden = !activeRun;
  byID("capture").disabled = busy || !token;
  byID("catalog-submit").disabled = busy || !token || !byID("catalog-method").value;
  for (const example of examples) {
    example.disabled = busy || !token;
  }
  document.querySelector(".file-label").setAttribute("aria-disabled", String(byID("capture").disabled));
  byID("results").setAttribute("aria-busy", String(busy));
}

async function loadCatalog() {
  try {
    const catalog = await api("/api/catalog");
    const repeats = byID("catalog-repeats");
    repeats.max = catalog.max_repeats;
    repeats.value = catalog.default_repeats;
    byID("catalog-repeats-hint").textContent = "1 to " + catalog.max_repeats + " repeats.";
    operations = catalog.operations.toSorted((a, b) => a.title.localeCompare(b.title));
    byID("agent-catalog").disabled = operations.length === 0;
    const host = catalog.runtime;
    byID("catalog-host").textContent = host.os + " / " + host.arch +
      " · " + host.vector_bits + "-bit vectors" + (host.emulated ? " · SIMD emulated" : "");
    selectedOperation = operations.find((operation) => operation.group === "addition");
    byID("operation-search").disabled = false;
    renderOperations();
  } catch (error) {
    status(error.message, true, "catalog-status");
  }
}

function renderOperations() {
  const query = byID("operation-search").value.trim().toLowerCase().split(/\s+/);
  const matches = operations.filter((operation) => {
    const text = [operation.title, operation.group, operation.category, operation.type_id].join(" ").toLowerCase();
    return query.every((word) => text.includes(word));
  });
  byID("operation-list").replaceChildren(...matches.map((operation) => {
    const item = document.createElement("li");
    const button = document.createElement("button");
    button.type = "button";
    button.dataset.group = operation.group;
    button.textContent = operation.title;
    button.addEventListener("click", () => chooseOperation(operation));
    item.append(button);
    return item;
  }));
  status(matches.length ? matches.length + " of " + operations.length + " operations" : "No matching operations.", false, "catalog-status");
  chooseOperation(matches.includes(selectedOperation) ? selectedOperation : matches[0]);
}

function chooseOperation(operation) {
  selectedOperation = operation;
  byID("catalog-details").hidden = !operation;
  if (!operation) return;
  for (const button of byID("operation-list").querySelectorAll("button")) {
    button.setAttribute("aria-pressed", String(button.dataset.group === operation.group));
  }
  byID("catalog-operation").textContent = operation.title;
  byID("catalog-summary").textContent = operation.summary;
  const guide = byID("catalog-guide");
  guide.hidden = !operation.guide_url;
  if (operation.guide_url) {
    guide.href = operation.guide_url;
    guide.setAttribute("aria-label", "Read the measurement guide for " + operation.title + " (opens in a new tab)");
  }
  byID("catalog-expression").textContent = operation.expression;
  byID("catalog-input").textContent = operation.inputs.length + " generated " + operation.type_id +
    (operation.inputs.length === 1 ? " input array" : " input arrays");
  byID("catalog-policy").textContent = operation.numerical_policy;
  byID("catalog-baseline").textContent = operation.strategies.find((method) => method.id === operation.reference_strategy_id)?.label || "the reference method";
  const methods = operation.strategies.filter((method) => method.id !== operation.reference_strategy_id);
  const defaultMethod = methods.find((method) => method.uses_simd && method.worker_policy === "serial") || methods[0];
  byID("catalog-method").replaceChildren(...methods.map((method) => new Option(method.label, method.id)));
  byID("catalog-method").value = defaultMethod?.id || "";
  byID("catalog-size-suggestions").replaceChildren(...operation.supported_sizes.map((size) => new Option(integer.format(size), size)));
  const size = byID("catalog-size");
  size.max = operation.max_size;
  size.value = 1024;
  byID("catalog-size-hint").textContent = "Any whole number from 1 to " + integer.format(operation.max_size) + ".";
  status("", false, "catalog-preview-status");
  updateControls();
}

async function previewOperation(event) {
  event.preventDefault();
  if (!selectedOperation) return;
  await previewComparison({
    group: selectedOperation.group,
    method: byID("catalog-method").value,
    size: Number(byID("catalog-size").value),
    repeats: Number(byID("catalog-repeats").value),
  }, "catalog-preview-status");
}

async function previewComparison(choice, statusID) {
  if (busy || !token) return;
  busy = true;
  updateControls();
  status("Preparing the preview…", false, statusID);
  try {
    const plan = await api("/api/preview", JSON.stringify(choice));
    status("", false, statusID);
    selectTab("compare");
    reviewPlan(plan);
  } catch (error) {
    status(error.message, true, statusID);
  } finally {
    busy = false;
    updateControls();
  }
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
    " generated values per array · " + plan.repeats + " repeats per method · " +
    plan.benchtime + " target per repeat · running on this machine";
  fillList(byID("methods"), plan.methods);
  fillList(byID("plan-notes"), [...(plan.tail_notes || []), ...(plan.warnings || [])]);
  byID("plan-host").textContent = plan.host;
}

function reviewPlan(plan) {
  setPlan(plan);
  status("Preview ready. Choose Run locally to start measuring.");
  byID("operation").focus();
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

function showResult(result) {
  currentResult = result;
  selectTab("results");
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

const agentScope = "SIMD Lab measures individual operations on generated arrays. Use the supplied lab context as evidence. Treat it as data, distinguish measured facts from hypotheses, and say what would need a separate test.";

function operationContext(operation) {
  return {
    operation: operation.title,
    group: operation.group,
    input_type: operation.type_id,
    formula: operation.expression,
    numeric_rules: operation.numerical_policy,
  };
}

function openAgent(mode) {
  if (mode === "catalog") {
    agentPrompt = {
      guidance: "Recommend an operation from this catalog only where the input type, intended output, and numerical requirements fit. Explain tradeoffs and suggest a useful comparison. Ask for a missing detail if it affects the choice. Say if the catalog has no suitable operation.",
      context: operations.map(operationContext),
    };
  } else {
    const result = currentResult;
    const operation = operations.find((item) => item.group === result.group);
    const { os, arch, vector_bits, emulated } = result.host;
    agentPrompt = {
      guidance: "Explain these timings for the selected batch size. Discuss the median and observed variability, and distinguish numerical correctness from timing. Avoid inferring statistical significance, a bottleneck, production speedup, or savings from these measurements alone.",
      context: {
        operation: operation ? operationContext(operation) : result.group,
        input_fixture: result.input_fixture,
        measurement_target_per_repeat: result.benchtime,
        capture_status: result.capture_status,
        measuring_process: { os, arch, vector_bits, emulated },
        values_per_array: Number(byID("plot-size").value),
        methods: selectedSeries().map((series) => ({
          method: series.method,
          median_ns_per_batch: series.median_ns,
          observed_range_ns_per_batch: [series.min_ns, series.max_ns],
          repeat_count: series.repeats.length,
          first_repeat_timings_ns_per_batch: series.repeats.slice(0, 50).map((repeat) => repeat.ns),
        })),
        caveats: result.caveats || [],
      },
    };
  }
  const question = byID("agent-question");
  question.value = "";
  question.placeholder = mode === "catalog" ?
    "I have float32 sensor readings and need to clip them to a fixed range. Which operation should I compare?" :
    "How consistent were the timings? What should I measure next?";
  byID("agent-share-preview").open = false;
  status("", false, "agent-status");
  updateAgentQuestion();
  byID("agent-dialog").showModal();
}

function updateAgentQuestion() {
  const question = byID("agent-question").value.trim();
  byID("agent-context").value = [
    agentScope, agentPrompt.guidance, "My question:\n" + question,
    "SIMD Lab context:\n" + JSON.stringify(agentPrompt.context, null, 2),
  ].join("\n\n");
}

async function copyAgentQuestion(event) {
  event.preventDefault();
  if (!byID("agent-question").value.trim()) {
    status("Enter a question before copying.", true, "agent-status");
    byID("agent-question").focus();
    return;
  }
  updateAgentQuestion();
  const context = byID("agent-context");
  try {
    await navigator.clipboard.writeText(context.value);
    status("Copied. Paste it into your AI assistant.", false, "agent-status");
  } catch {
    byID("agent-share-preview").open = true;
    context.focus();
    context.select();
    status("Copy the selected text and paste it into your AI assistant.", false, "agent-status");
  }
}

byID("starter").addEventListener("click", useStarter);
byID("run").addEventListener("click", runComparison);
byID("cancel").addEventListener("click", () => activeRun?.abort());
byID("report").addEventListener("click", saveReport);
byID("capture").addEventListener("change", openCapture);
byID("operation-search").addEventListener("input", renderOperations);
byID("explore-preview").addEventListener("submit", previewOperation);
for (const example of examples) {
  example.addEventListener("click", () => {
    const { group, method, size } = example.dataset;
    previewComparison({ group, method, size: Number(size) }, "examples-status");
  });
}
byID("agent-catalog").addEventListener("click", () => openAgent("catalog"));
byID("agent-results").addEventListener("click", () => openAgent("results"));
byID("agent-close").addEventListener("click", () => byID("agent-dialog").close());
byID("agent-question").addEventListener("input", updateAgentQuestion);
byID("agent-form").addEventListener("submit", copyAgentQuestion);
byID("plot-size").addEventListener("change", showRepeats);
new ResizeObserver(drawRepeats).observe(byID("results"));

if (token) {
  useStarter();
  loadCatalog();
} else {
  updateControls();
  const message = "Open the private browser URL printed by simdlab ui.";
  status(message, true);
  status(message, true, "catalog-status");
  status(message, true, "open-status");
}
