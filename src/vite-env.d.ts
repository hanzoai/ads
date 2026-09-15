/// <reference types="vite/client" />

// A stylesheet is not a module with a type; these imports exist for their
// effect. `@hanzo/font/css` ships the woff2 and the `@font-face` that names it,
// and has no declaration of its own.
declare module '@hanzo/font/css'
declare module '*.css'
