// HyperFleet Commit Message Standard
// Format: HYPERFLEET-XXX - <type>: <subject>
// See: https://github.com/openshift-hyperfleet/architecture/blob/main/hyperfleet/standards/commit-standard.md

module.exports = {
  extends: ["@commitlint/config-conventional"],
  parserPreset: {
    parserOpts: {
      // Support both formats:
      //   HYPERFLEET-123 - feat: add new feature
      //   feat: add new feature
      headerPattern: /^(?:HYPERFLEET-\d+\s*-\s*)?(\w+)(?:\((.+)\))?:\s(.+)$/,
      headerCorrespondence: ["type", "scope", "subject"],
    },
  },
  plugins: [
    {
      rules: {
        "header-max-length-excluding-jira": (parsed, _when, value) => {
          const header = parsed.header || "";
          const withoutJira = header.replace(/^HYPERFLEET-\d+\s*-\s*/, "");
          const valid = withoutJira.length <= value;
          return [
            valid,
            `header (excluding JIRA prefix) must not be longer than ${value} characters, current length is ${withoutJira.length}`,
          ];
        },
      },
    },
  ],
  rules: {
    // Disable default header-max-length in favor of custom rule that excludes JIRA prefix
    "header-max-length": [0],
    "header-max-length-excluding-jira": [2, "always", 72],
    "type-enum": [
      2,
      "always",
      [
        "feat",
        "fix",
        "docs",
        "style",
        "refactor",
        "perf",
        "test",
        "build",
        "ci",
        "chore",
        "revert",
      ],
    ],
    // Body has no hard limit per commit-standard.md
    "body-max-line-length": [0],
    "footer-max-line-length": [0],
  },
};
