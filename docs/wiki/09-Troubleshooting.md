# Troubleshooting

## The browser did not open

Keep the terminal running and copy the private address printed by `simdlab ui`. If the session expired, restart the lab and open the new address.

## The preview expired

Use **Explore** to preview the same settings again, or reload the addition starter in **Compare**. Review the new preview before choosing **Run locally**.

## A file will not open

**Results** opens a lab measurement capture. Choose the original measurement JSON instead of the formatted Markdown or CSV report. See [Reports and evidence](05-Reports-and-evidence).

## A percentage is missing

A timing difference needs a compatible scalar observation at the same operation and batch size. Compare both methods under the same fixture and host conditions.

## The vector method took longer

Check batch size, the remainder, and the repeat spread. Setup or tail work can dominate a small batch. If the method uses parallel workers, coordination can add time. See [Understanding results](06-Understanding-results).

## A fused operation reports software execution

Read the recorded path as part of the result. A software fused method and a vector fused method can use the same one-rounding rule with different costs. See [Fused affine transforms and rounding](17-Fused-affine-and-rounding).

## AI chat setup and model problems

Local chat is experimental. If Ollama is unavailable, start it and use **Refresh**.
If a model is missing, choose an installed Chat model or consult the setup guide
before installing one. Use **Stop** for a slow reply. If there is no usable draft,
create the comparison in Explore or copy the question and context to another
assistant. Changing Chat model starts a new thread. CLI previews require explicit
`-model MODEL`; ordinary guided starter questions need no model.

[AI setup, tested models, and API-key instructions](34-AI-setup-and-model-recommendations).
