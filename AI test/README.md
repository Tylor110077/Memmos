# Aliyun Bailian streaming chat test

This folder contains terminal REPL examples for Aliyun Bailian's OpenAI-compatible chat API with streaming output and persisted multi-turn history.

## Recommended: Go version

```bash
export ALIYUN_BAILIAN_API_KEY=your_api_key
go run './AI test/aliyun_stream_chat.go'
```

- Reads the API key from `ALIYUN_BAILIAN_API_KEY`
- Falls back to `docs/api.md` if the env var is not set
- Uses `qwen3.5-plus` by default
- Sends `enable_thinking=false` by default so the stream is direct answer output
- Calls the Beijing OpenAI-compatible endpoint at `https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions`
- Streams assistant tokens to the terminal
- Stores conversation history in `AI test/chat-history-go.json`

## Commands

- `/history` prints the current message array
- `/clear` resets the stored conversation to the system prompt
- `/exit` quits

## Optional environment variables

```bash
ALIYUN_BAILIAN_API_KEY=your_api_key go run './AI test/aliyun_stream_chat.go'
ALIYUN_BAILIAN_MODEL=qwen3.5-plus go run './AI test/aliyun_stream_chat.go'
ALIYUN_BAILIAN_ENABLE_THINKING=true go run './AI test/aliyun_stream_chat.go'
ALIYUN_BAILIAN_SYSTEM_PROMPT="You are a domain expert." go run './AI test/aliyun_stream_chat.go'
ALIYUN_BAILIAN_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1 go run './AI test/aliyun_stream_chat.go'
```

## Legacy Node version

```bash
npm run ai:test
```

## `docs/api.md` format

If you choose to use `docs/api.md`, keep it as a local helper file with one non-comment line containing the key:

```text
# Local-only fallback for Aliyun Bailian API key
<replace-with-your-api-key>
```
