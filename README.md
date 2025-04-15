# AgentAPI

Control Claude Code, Goose, and Aider with an HTTP API.
<div>
  <img width="800px" src="https://github.com/user-attachments/assets/4b9950f5-bfa7-4f38-bf09-58364f78c100">
</div>

## Usage

AgentAPI is under active development. Release binaries will be available soon.

To run the server with Claude Code:

```bash
go run main.go server -- claude
```

You may also pass additional arguments to the agent command. For example:

```bash
go run main.go server -- claude --allowedTools "Bash(git*) Edit Replace"
```

By default, the server runs on port 3284. You can inspect the available endpoints by opening http://localhost:3284/docs.

There are 4 endpoints:

- GET `/messages` - returns a list of all messages in the conversation with the agent
- POST `/message` - sends a message to the agent. When a 200 response is returned, AgentAPI has detected that the agent started processing the message
- GET `/status` - returns the current status of the agent, either "waiting_for_input" or "running"
- GET `/events` - an SSE stream of events from the agent: new messages and status updates

## How it works

AgentAPI runs an in-memory terminal emulator. It translates API calls into appropriate terminal keystrokes, and parses the agent's outputs into individual messages using a set of heuristics.
