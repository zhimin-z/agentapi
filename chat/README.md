# AgentAPI Chat Interface

A simple ChatGPT-like interface for AgentAPI.

## Setup

1. Ensure the AgentAPI backend server is running on `localhost:8080`

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
   ```bash
   npm run dev
   ```

4. Open [http://localhost:3000](http://localhost:3000) in your browser

## Features

- Real-time message polling (every 1 second)
- Server status monitoring
- Simple chat interface
- Responsive design

## API Endpoints Used

- `GET /messages` - Retrieves all messages
- `POST /message` - Sends a new message
- `GET /status` - Checks server status