#!/usr/bin/env node
// Wrapper script for commitlint that uses the bundled HyperFleet config.
// This is necessary because pre-commit runs hooks from the consuming repo's
// working directory, but we need commitlint to use this package's config.

const path = require("path");
const { execFileSync } = require("child_process");

const configPath = path.resolve(__dirname, "..", "commitlint.config.js");
const commitMsgFile = process.argv[2];

if (!commitMsgFile) {
  console.error("Error: commit message file path is required");
  process.exit(1);
}

const commitlintBin = path.resolve(
  __dirname,
  "..",
  "node_modules",
  ".bin",
  "commitlint",
);

try {
  execFileSync(
    commitlintBin,
    ["--edit", commitMsgFile, "--config", configPath],
    {
      stdio: "inherit",
    },
  );
} catch (e) {
  process.exit(e.status || 1);
}
