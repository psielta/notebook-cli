#!/usr/bin/env node
// Publish every package assembled in dist-npm/ to the npm registry, in the
// correct order for the optionalDependencies pattern: all platform packages
// FIRST, then the main package LAST (so its optionalDependencies resolve).
//
// Idempotent: a version already on the registry is skipped, so a re-run after a
// partial failure finishes the remaining packages instead of erroring with 403.
//
// Usage:
//   node scripts/publish-npm.mjs                 # publish (used locally)
//   node scripts/publish-npm.mjs --provenance    # CI: attach build provenance
//   node scripts/publish-npm.mjs --dry-run       # print what would happen
//
// Requires: `node scripts/build-npm.mjs` to have populated dist-npm/ first.
// In CI: needs `permissions: id-token: write` and NODE_AUTH_TOKEN in the env
// (setup-node with registry-url writes the .npmrc that references it).

import { execFileSync } from "node:child_process";
import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(__dirname, "..");
const outDir = join(repoRoot, "dist-npm");

const SCOPE = "@psielta";
const provenance = process.argv.includes("--provenance");
const dryRun = process.argv.includes("--dry-run");
const allowDev = process.argv.includes("--allow-dev");

if (!existsSync(outDir)) {
  console.error(`dist-npm/ not found. Run: node scripts/build-npm.mjs first.`);
  process.exit(1);
}

// Collect every package dir (each contains a package.json) under dist-npm/.
function findPackageDirs() {
  const scopeDir = join(outDir, SCOPE);
  if (!existsSync(scopeDir)) return [];
  return readdirSync(scopeDir, { withFileTypes: true })
    .filter((d) => d.isDirectory())
    .map((d) => join(scopeDir, d.name))
    .filter((dir) => existsSync(join(dir, "package.json")));
}

const dirs = findPackageDirs();
if (dirs.length === 0) {
  console.error(`No packages found under ${join(outDir, SCOPE)}.`);
  process.exit(1);
}

// Order: platform packages (name contains "/nb-") first, main package last.
const pkgs = dirs
  .map((dir) => ({ dir, manifest: JSON.parse(readFileSync(join(dir, "package.json"), "utf8")) }))
  .sort((a, b) => {
    const am = a.manifest.name === `${SCOPE}/notebook-cli` ? 1 : 0;
    const bm = b.manifest.name === `${SCOPE}/notebook-cli` ? 1 : 0;
    return am - bm; // main (1) sorts last
  });

// Refuse to publish a local/dev build (e.g. the 0.0.0-dev fallback or a
// -test/SNAPSHOT version) to the public registry by accident. Override with
// --allow-dev for an intentional prerelease.
const version = pkgs[0].manifest.version;
function isReleaseVersion(v) {
  if (!v || v.startsWith("0.0.0")) return false;
  if (/-(dev|test|snapshot|local|dirty)\b/i.test(v)) return false;
  return /^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/.test(v);
}
if (!dryRun && !allowDev && !isReleaseVersion(version)) {
  console.error(
    `Refusing to publish non-release version "${version}".\n` +
      `Set a real version via a git tag or --version on build-npm.mjs, or pass\n` +
      `--allow-dev to publish this version intentionally (or --dry-run to preview).`
  );
  process.exit(1);
}

// The main package's bin/nb is a shebang script: CRLF endings break *nix shells
// ("bad interpreter"). npm fixes its exec bit on install, but not line endings.
function checkLauncherEndings({ dir, manifest }) {
  if (manifest.name !== `${SCOPE}/notebook-cli`) return;
  const launcher = join(dir, "bin", "nb");
  if (existsSync(launcher) && readFileSync(launcher, "utf8").includes("\r")) {
    console.error(
      `${manifest.name}: bin/nb has CRLF line endings, which break the shebang on\n` +
        `       Linux/macOS. Ensure .gitattributes pins it to LF and rebuild.`
    );
    if (!dryRun) process.exit(1);
  }
}

// Unix binaries must ship mode 0755 or they fail with EACCES on the user's
// machine. Windows can't set the exec bit, so a unix package cross-built there
// would publish broken. Guard against it: hard-fail on a real publish, warn on
// a dry run.
function checkExecBit({ dir, manifest }) {
  const os = manifest.os || [];
  if (os.length === 0 || os.includes("win32")) return; // main pkg / windows: n/a
  const binPath = join(dir, "bin", "nb");
  let executable = false;
  try {
    executable = (statSync(binPath).mode & 0o111) !== 0;
  } catch {
    /* missing binary handled below */
  }
  if (executable) return;
  const msg =
    `${manifest.name}: the unix binary at bin/nb is not marked executable, so the\n` +
    `       published package would fail with EACCES. Build AND publish from Linux/macOS\n` +
    `       (the release CI runs on ubuntu); Windows cannot set the executable bit.`;
  if (dryRun) {
    console.warn(`warn  ${msg}`);
  } else {
    console.error(`error ${msg}`);
    process.exit(1);
  }
}

function alreadyPublished(name, version) {
  try {
    const out = execFileSync("npm", ["view", `${name}@${version}`, "version"], {
      stdio: ["ignore", "pipe", "ignore"],
    })
      .toString()
      .trim();
    return out === version;
  } catch {
    return false; // 404 / not found -> not published yet
  }
}

let published = 0;
let skipped = 0;

for (const { dir, manifest } of pkgs) {
  const { name, version } = manifest;
  checkExecBit({ dir, manifest });
  checkLauncherEndings({ dir, manifest });
  if (alreadyPublished(name, version)) {
    console.log(`skip  ${name}@${version} (already on registry)`);
    skipped++;
    continue;
  }

  const args = ["publish", dir, "--access", "public"];
  if (provenance) args.push("--provenance");

  if (dryRun) {
    console.log(`DRY   npm ${args.join(" ")}`);
    continue;
  }

  console.log(`pub   ${name}@${version}`);
  try {
    execFileSync("npm", args, { stdio: "inherit" });
    published++;
  } catch (e) {
    // A racy re-run can lose the alreadyPublished check; if the version landed
    // in the meantime, treat it as published rather than failing the release.
    if (alreadyPublished(name, version)) {
      console.log(`skip  ${name}@${version} (already published during this run)`);
      skipped++;
    } else {
      throw e;
    }
  }
}

console.log(`\nDone. published=${published} skipped=${skipped}${dryRun ? " (dry-run)" : ""}`);
