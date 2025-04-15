# AgentAPI

Control Claude Code, Goose, and Aider with an HTTP API.

<div>
  <img width="800px" src="https://github.com/user-attachments/assets/4b9950f5-bfa7-4f38-bf09-58364f78c100">
</div>

## Quickstart

Install `agentapi` by downloading the latest release binary from the [releases page](https://github.com/coder/agentapi/releases) and putting it in your PATH.

Run a Claude Code server (assumes `claude` is installed on your system):

```bash
agentapi server -- claude
```

Send a message to the agent:

```bash
curl -X POST localhost:3284/message \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello, agent!", "type": "user"}'
```

Get the conversation history:

```bash
curl localhost:3284/messages
```

## CLI Commands

### `agentapi server`

Run an HTTP server that lets you control an agent. If you'd like to start an agent with additional arguments, pass the full agent command after the `--` flag.

```bash
agentapi server -- claude --allowedTools "Bash(git*) Edit Replace"
```

You may also use `agentapi` to run the Aider and Goose agents:

```bash
agentapi server -- aider --model sonnet --api-key anthropic=sk-ant-apio3-XXX 
agentapi server -- goose
```

By default, the server runs on port 3284. The server exposes an OpenAPI schema at http://localhost:3284/openapi.json. You may also inspect the available endpoints at http://localhost:3284/docs.

There are 4 endpoints:

- GET `/messages` - returns a list of all messages in the conversation with the agent
- POST `/message` - sends a message to the agent. When a 200 response is returned, AgentAPI has detected that the agent started processing the message
- GET `/status` - returns the current status of the agent, either "stable" or "running"
- GET `/events` - an SSE stream of events from the agent: message and status updates

### `agentapi attach`

Attach to a running agent's terminal session.

```bash
agentapi attach --url localhost:3284
```

Press `ctrl+c` to detach from the session.

## How it works

AgentAPI runs an in-memory terminal emulator. It translates API calls into appropriate terminal keystrokes, and parses the agent's outputs into individual messages. At the time of writing, none of the agents expose a native HTTP API. Once they do, AgentAPI will be updated to support them.
