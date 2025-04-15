# AgentAPI Chat Interface

A simple ChatGPT-like interface for AgentAPI. It's a demo showcasing how to use AgentAPI. 95% of the code was generated with Claude Code.

## Development Setup

1. Ensure the AgentAPI backend server is running on `localhost:3284`

2. Install dependencies:

   ```bash
   npm install
   ```

3. Start the development server:

   ```bash
   npm run dev
   ```

4. Open [http://localhost:3000](http://localhost:3000) in your browser

## Static Build for Hosting

This application can be built as a static site for deployment to any static web hosting service.

### Building the Static Site

```bash
# Generate the static export
npm run export
```

This will create a static build in the `out` directory, which can be deployed to any static hosting service.

### Testing the Static Build Locally

```bash
# Serve the static files locally
npm run serve-static
```

This will start a local server to test the static build.
