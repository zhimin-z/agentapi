/** @type {import('next').NextConfig} */
const isGitHubPages = process.env.GITHUB_PAGES === "true";
const repo = "agentapi";
const subPath = "chat"; // Subdirectory within the repo

const nextConfig = {
  // Enable static exports
  output: "export",

  // Disable image optimization since it's not supported in static exports
  images: {
    unoptimized: true,
  },

  // Configure base path for GitHub Pages (repo/chat)
  basePath: isGitHubPages ? `/${repo}/${subPath}` : `/${subPath}`,

  // Configure asset prefix for GitHub Pages - helps with static asset loading
  assetPrefix: isGitHubPages ? `/${repo}/${subPath}/` : `/${subPath}/`,

  // Configure trailing slashes (recommended for static exports)
  trailingSlash: true,
};

export default nextConfig;
