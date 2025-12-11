# Tunnel Web UI

Modern React frontend built with **Vite**, **React 19**, and **MUI v7**.

## Prerequisites

- Node.js 20+
- pnpm 9+

## Development

```bash
# Install dependencies
pnpm install

# Start dev server (hot reload)
pnpm dev

# Build for production
pnpm build

# Preview production build
pnpm preview

# Analyze bundle size
pnpm analyze
```

## Environment Variables

- `API_URL` - Backend API URL (default: `https://vu-de-2.vpnhouse.net`)

**Example:**

```bash
API_URL=https://your-api.com pnpm dev
```

## Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| React | 19 | UI Framework |
| Vite | 7 | Build Tool |
| MUI | 7 | Component Library |
| React Router | 7 | Routing |
| Effector | 23 | State Management |
| TypeScript | 5 | Type Safety |

## Build Output

The build is optimized with:

- **Code Splitting** - Vendor chunks for better caching
- **Brotli Compression** - ~70% smaller than gzip
- **Tree Shaking** - Only used code is bundled
- **CSS Code Splitting** - Per-route CSS loading

Typical bundle sizes (Brotli compressed):

| Chunk | Size |
|-------|------|
| MUI Core | ~79 KB |
| App Code | ~72 KB |
| Date Pickers | ~52 KB |
| React | ~14 KB |
| Effector | ~9 KB |
| Utilities | ~6 KB |
