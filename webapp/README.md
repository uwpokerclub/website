# UWPSC Web Application

> **📖 For complete documentation, please refer to the main [README.md](../README.md)**

This directory contains the React TypeScript frontend for the UWPSC Website.

## Quick Start

From the webapp directory:
```bash
npm install
npm run dev
```

## Key Information

- **Port**: 5173
- **Technology**: React + TypeScript + Vite
- **Testing**: `npm test`
- **Build**: `npm run build`

## Directory Structure

- `public/` - Static assets (favicon, images, etc.)
- `src/` - Source code
  - `components/` - Reusable React components
  - `pages/` - Route-level page components
  - `hooks/` - Custom React hooks
  - `contexts/` - React context providers
  - `sdk/` - API client for backend communication
  - `types/` - TypeScript type definitions
  - `utils/` - Utility functions

## Development

**Local Development:**
```bash
npm install
npm run dev
```

**Testing:**
```bash
npm test              # Run tests
npm run test:watch    # Interactive testing
```

**Production Build:**
```bash
npm run build
```

For complete setup instructions, database management, and deployment information, see the main [README.md](../README.md).

## License

Licensed under the Apache 2.0 License - see the [LICENSE](../LICENSE) file for details.
