import fs from "node:fs/promises";
import path from "node:path";
import readline from "node:readline/promises";
import { stdin as input, stdout as output } from "node:process";

const endpoint =
  process.env.ALIYUN_BAILIAN_BASE_URL ??
  "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions";
const model = process.env.ALIYUN_BAILIAN_MODEL ?? "qwen-plus";
const envApiKey = process.env.ALIYUN_BAILIAN_API_KEY?.trim();
const cwd = process.cwd();
const apiKeyPath = path.resolve(cwd, "docs/api.md");
const historyPath = path.resolve(cwd, "AI test/chat-history.json");
const systemPrompt =
  process.env.ALIYUN_BAILIAN_SYSTEM_PROMPT ??
  "You are a concise and helpful AI assistant.";

async function ensureHistoryFile() {
  try {
    await fs.access(historyPath);
  } catch {
    const initialHistory = [{ role: "system", content: systemPrompt }];
    await fs.writeFile(historyPath, JSON.stringify(initialHistory, null, 2) + "\n");
  }
}

async function loadApiKey() {
  if (envApiKey) {
    return envApiKey;
  }

  const raw = await fs.readFile(apiKeyPath, "utf8");
  const key = raw
    .split(/\r?\n/)
    .map((line) => line.trim())
    .find((line) => line && !line.startsWith("#") && !line.startsWith("<"));
  if (!key) {
    throw new Error(`No API key found. Set ALIYUN_BAILIAN_API_KEY or update ${apiKeyPath}.`);
  }
  return key;
}

async function loadHistory() {
  await ensureHistoryFile();
  const raw = await fs.readFile(historyPath, "utf8");
  const parsed = JSON.parse(raw);
  if (!Array.isArray(parsed) || parsed.length === 0) {
    return [{ role: "system", content: systemPrompt }];
  }
  return parsed;
}

async function saveHistory(messages) {
  await fs.writeFile(historyPath, JSON.stringify(messages, null, 2) + "\n");
}

async function resetHistory() {
  const messages = [{ role: "system", content: systemPrompt }];
  await saveHistory(messages);
  return messages;
}

async function streamChatCompletion({ apiKey, messages }) {
  const response = await fetch(endpoint, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${apiKey}`,
    },
    body: JSON.stringify({
      model,
      stream: true,
      messages,
    }),
  });

  if (!response.ok || !response.body) {
    const errorText = await response.text();
    throw new Error(`Request failed (${response.status}): ${errorText}`);
  }

  const decoder = new TextDecoder("utf8");
  const reader = response.body.getReader();
  let buffer = "";
  let assistantReply = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }

    buffer += decoder.decode(value, { stream: true });
    const events = buffer.split("\n\n");
    buffer = events.pop() ?? "";

    for (const event of events) {
      const lines = event
        .split("\n")
        .map((line) => line.trim())
        .filter(Boolean);

      for (const line of lines) {
        if (!line.startsWith("data:")) {
          continue;
        }

        const data = line.slice(5).trim();
        if (data === "[DONE]") {
          return assistantReply.trim();
        }

        const payload = JSON.parse(data);
        const delta = payload.choices?.[0]?.delta?.content;
        if (typeof delta === "string" && delta.length > 0) {
          assistantReply += delta;
          output.write(delta);
        }
      }
    }
  }

  return assistantReply.trim();
}

async function main() {
  const apiKey = await loadApiKey();
  let history = await loadHistory();
  const rl = readline.createInterface({ input, output });

  output.write(`Aliyun Bailian streaming chat ready.\n`);
  output.write(`Model: ${model}\n`);
  output.write(`History file: ${historyPath}\n`);
  output.write(`Commands: /exit, /clear, /history\n\n`);

  while (true) {
    let prompt;
    try {
      prompt = await rl.question("You: ");
    } catch (error) {
      if (error?.code === "ERR_USE_AFTER_CLOSE") {
        break;
      }
      throw error;
    }

    const text = prompt.trim();

    if (!text) {
      continue;
    }

    if (text === "/exit") {
      break;
    }

    if (text === "/clear") {
      history = await resetHistory();
      output.write("History cleared.\n\n");
      continue;
    }

    if (text === "/history") {
      output.write(JSON.stringify(history, null, 2) + "\n\n");
      continue;
    }

    history.push({ role: "user", content: text });
    output.write("Assistant: ");

    try {
      const reply = await streamChatCompletion({ apiKey, messages: history });
      output.write("\n\n");
      history.push({ role: "assistant", content: reply });
      await saveHistory(history);
    } catch (error) {
      history.pop();
      output.write(`\n\nRequest error: ${error.message}\n\n`);
    }
  }

  rl.close();
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
