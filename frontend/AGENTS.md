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

