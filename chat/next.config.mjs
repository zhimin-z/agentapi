/** @type {import('next').NextConfig} */
const nextConfig = {
  // Enable static exports
  output: 'export',
  
  // Disable image optimization since it's not supported in static exports
  images: {
    unoptimized: true,
  },
  
  // Remove headers config for static export
};

export default nextConfig;