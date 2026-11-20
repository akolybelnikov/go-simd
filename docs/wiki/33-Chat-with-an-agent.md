# Chat with an agent

Use chat to learn about SIMD, find an operation for your workload, or discuss
a saved result. Start with one of these questions.

| Where to open chat | A question to try |
| --- | --- |
| **Explore → Ask an agent about your use case** | “I have float32 sensor readings and want to clip them between two bounds. Which operation should I compare?” |
| **Results → Ask an agent about these results** | “Were these repeats consistent? What comparison would help me understand the difference?” |

In Results, choose the batch size first if the capture contains several.
The answer uses the context for that selected batch.

## Choose how to ask

The panel opens on **Copy to another assistant**. Write your question,
review what will be copied, and paste it into the assistant you already
use. Preparing the context works offline. **Remote · OpenAI** is an
optional connection.

### OpenAI

Set `OPENAI_API_KEY` in the terminal that launches `simdlab`. Replace the
example text with your own API key.

**Windows PowerShell**

```powershell
$env:OPENAI_API_KEY = 'your-api-key'
simdlab ui
```

**macOS or Linux terminal**

```sh
export OPENAI_API_KEY='your-api-key'
simdlab ui
```

Select **Remote · OpenAI · gpt-6.1-sol** in **How to ask**. This app currently uses
that model for OpenAI chat. Requests use your API account, and this lab
session allows up to 24 OpenAI requests. [OpenAI API setup](https://developers.openai.com/api/docs/quickstart)

## Send a question and follow up

Write your question, expand **Review what will be sent** to inspect the
context, then choose **Send**. Continue with a follow-up in the same panel.
The lab includes up to three recent exchanges with the next question.
In copy mode, the preview is labeled **Review what will be copied**.

The shared context includes catalog information and, in Results, the
selected batch’s measured timings, host details, and caveats. File paths
and raw arrays are omitted. Your typed question is also sent to the chosen
provider. Review it before sharing private data.

**Stop** cancels the pending request. **New chat** clears the conversation.
Changing provider or closing and reopening the panel starts a new thread.
If a reply fails or takes too long, you can copy the question and context.

## Try a suggested comparison

If the answer offers **Edit suggested comparison**, choose it to fill in
Explore. Review or change the method, batch size, and repeat count. Choose
**Preview comparison**, then **Run locally** to approve the measurement.
The experiment uses generated inputs.

Keep the measured evidence and caveats alongside the explanation. A proposed
cache or bandwidth explanation is a hypothesis you can test with another
comparison. [Understanding results](06-Understanding-results)
