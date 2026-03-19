const fs = require("fs");
const path = require("path");

const SHARD_COUNT = parseInt(process.env.SHARD_COUNT || "3", 10);
const SPEC_DIR = path.resolve(__dirname, "../../cypress/e2e");

// Count `it(` occurrences in each spec file
const specs = fs
  .readdirSync(SPEC_DIR)
  .filter((f) => f.endsWith(".cy.ts"))
  .map((f) => {
    const content = fs.readFileSync(path.join(SPEC_DIR, f), "utf8");
    const count = (content.match(/\bit\(/g) || []).length;
    return { file: `cypress/e2e/${f}`, count };
  })
  .sort((a, b) => b.count - a.count);

// Greedy bin-packing: assign each spec to the shard with the fewest tests
const shards = Array.from({ length: SHARD_COUNT }, (_, i) => ({
  shard: i + 1,
  specs: [],
  total: 0,
}));

for (const spec of specs) {
  const smallest = shards.reduce((min, s) => (s.total < min.total ? s : min));
  smallest.specs.push(spec.file);
  smallest.total += spec.count;
}

// Log assignments
for (const s of shards) {
  console.log(`Shard ${s.shard} (${s.total} tests): ${s.specs.join(", ")}`);
}

// Output matrix for GitHub Actions
const matrix = {
  include: shards.map((s) => ({
    shard: s.shard,
    specs: s.specs.join(","),
  })),
};

const output = `matrix=${JSON.stringify(matrix)}`;

if (process.env.GITHUB_OUTPUT) {
  fs.appendFileSync(process.env.GITHUB_OUTPUT, output + "\n");
} else {
  console.log(`\n${output}`);
}
