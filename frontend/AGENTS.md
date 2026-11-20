# Frontend Guidelines

## Project Structure & Module Organization

This is a Next.js App Router application. `app/page.tsx` is the home page, `app/layout.tsx` defines the root layout, and `next.config.ts` holds Next.js configuration. Place assets in `public/` when needed.

## Build, Test, and Development Commands

- `npm ci` installs the locked dependencies.
- `npm run dev` starts the Next.js development server.
- `npm run build` creates a production build; `npm start` serves it after building.

## Coding Style & Naming Conventions

Follow the existing TypeScript and TSX style: two-space indentation, double-quoted imports, and PascalCase React component names. Keep TypeScript compatible with the strict settings in `tsconfig.json`. No dedicated lint or formatting scripts are currently configured.

## Testing Guidelines

No frontend test framework or coverage threshold is configured yet; add focused tests when introducing a suitable framework, and verify frontend changes with a production build.


<!-- BEGIN:nextjs-agent-rules -->

## This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->
