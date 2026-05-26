module.exports = {
  // Run ESLint (auto-fix) + Prettier on staged Vue/TS files
  'frontend/src/**/*.{ts,vue}': (files) => [
    `pnpm --dir frontend exec eslint --fix ${files.join(' ')}`,
    `pnpm --dir frontend exec prettier --write ${files.join(' ')}`,
  ],
  // Run full golangci-lint on any staged Go change
  'backend/**/*.go': () => 'make lint-be',
}
