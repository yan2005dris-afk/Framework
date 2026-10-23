export default {
  paths: ["tests/e2e/features/**/*.feature"],
  import: ["tests/e2e/steps/**/*.ts"],
  format: ["summary"],
  formatOptions: { snippetInterface: "async-await" },
};
