#!/usr/bin/env node
// Build the Go `nb` binary for every supported platform and assemble the npm
// packages (one main package + one optional dependency per platform), ready to
// publish with scripts/publish-npm.mjs.
//
// Usage:
//   node scripts/build-npm.mjs --version 1.2.3
//   node scripts/build-npm.mjs                 # version from latest git tag, else 0.0.0-dev
//
// Output (clean-built each run):
//   dist-npm/@psielta/notebook-cli/            main package (JS launcher + package.json)
//   dist-npm/@psielta/nb-<os>-<arch>/          one package per target, each holding the binary
//
// The matching optional dependency is the only one npm installs on a given host
// (selected via the "os"/"cpu" fields), so there is no postinstall and no
// download at install time.
//
// IMPORTANT — executable bit: npm packs files with their on-disk mode. Unix
// binaries must be mode 0755 in the published tarball or `nb` fails with EACCES
// on the user's machine. Windows can't set that bit, so the OFFICIAL release
// must be built and published from Linux/macOS (the release CI runs on ubuntu).
// Building on Windows is fine for local testing of the host-platform package.

import { execFileSync } from "node:child_process";
import {
  chmodSync,
  cpSync,
  existsSync,
  mkdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(__dirname, "..");
const outDir = join(repoRoot, "dist-npm");
const mainSrcDir = join(repoRoot, "npm", "notebook-cli");

// --- Shared metadata (single source of truth) -------------------------------
// SCOPE must match the SCOPE in npm/notebook-cli/bin/nb (the launcher).
const SCOPE = "@psielta";
const VERSION_SYMBOL = "notebook-cli/internal/version.Version";
const REPO_URL = "git+https://github.com/psielta/notebook-cli.git";
const HOMEPAGE = "https://github.com/psielta/notebook-cli#readme";
const BUGS_URL = "https://github.com/psielta/notebook-cli/issues";
const LICENSE = "MIT";
const AUTHOR = "psielta";

// process.platform / process.arch use these exact strings, so the launcher can
// build the package name as `${SCOPE}/nb-${process.platform}-${process.arch}`.
const TARGETS = [
  { goos: "linux", goarch: "amd64", os: "linux", cpu: "x64" },
  { goos: "linux", goarch: "arm64", os: "linux", cpu: "arm64" },
  { goos: "darwin", goarch: "amd64", os: "darwin", cpu: "x64" },
  { goos: "darwin", goarch: "arm64", os: "darwin", cpu: "arm64" },
  { goos: "windows", goarch: "amd64", os: "win32", cpu: "x64" },
  { goos: "windows", goarch: "arm64", os: "win32", cpu: "arm64" },
];

// --- Version resolution ------------------------------------------------------
function normalizeVersion(v) {
  return String(v).trim().replace(/^v/, "");
}

function resolveVersion() {
  const i = process.argv.indexOf("--version");
  if (i !== -1 && process.argv[i + 1]) return normalizeVersion(process.argv[i + 1]);
  if (process.env.NB_VERSION) return normalizeVersion(process.env.NB_VERSION);
  try {
    const tag = execFileSync("git", ["describe", "--tags", "--abbrev=0"], {
      cwd: repoRoot,
    })
      .toString()
      .trim();
    if (tag) return normalizeVersion(tag);
  } catch {
    /* no tags yet */
  }
  return "0.0.0-dev";
}

const version = resolveVersion();
const goVersion = `v${version}`; // matches the existing `nb --version` convention

// --- Helpers -----------------------------------------------------------------
function writeJson(file, obj) {
  mkdirSync(dirname(file), { recursive: true });
  writeFileSync(file, JSON.stringify(obj, null, 2) + "\n");
}

function platformPkgName(t) {
  return `${SCOPE}/nb-${t.os}-${t.cpu}`;
}

function buildBinary(t, destBinPath) {
  const env = { ...process.env, GOOS: t.goos, GOARCH: t.goarch, CGO_ENABLED: "0" };
  execFileSync(
    "go",
    [
      "build",
      "-trimpath",
      "-ldflags",
      `-s -w -X ${VERSION_SYMBOL}=${goVersion}`,
      "-o",
      destBinPath,
      "./cmd/nb",
    ],
    { cwd: repoRoot, env, stdio: "inherit" }
  );
  // Ensure the unix exec bit is set (no-op / harmless on Windows hosts).
  if (t.goos !== "windows") chmodSync(destBinPath, 0o755);
}

// --- Build -------------------------------------------------------------------
console.log(`Building npm packages for ${SCOPE}/notebook-cli@${version}\n`);

if (existsSync(outDir)) rmSync(outDir, { recursive: true, force: true });
mkdirSync(outDir, { recursive: true });

const optionalDependencies = {};

for (const t of TARGETS) {
  const name = platformPkgName(t);
  const pkgDir = join(outDir, name);
  const binFile = t.goos === "windows" ? "nb.exe" : "nb";
  const binPath = join(pkgDir, "bin", binFile);

  console.log(`  • ${name} (${t.goos}/${t.goarch})`);
  mkdirSync(dirname(binPath), { recursive: true });
  buildBinary(t, binPath);

  writeJson(join(pkgDir, "package.json"), {
    name,
    version,
    description: `The ${t.os}-${t.cpu} binary for nb (notebook-cli).`,
    license: LICENSE,
    author: AUTHOR,
    homepage: HOMEPAGE,
    repository: { type: "git", url: REPO_URL },
    bugs: { url: BUGS_URL },
    os: [t.os],
    cpu: [t.cpu],
    files: ["bin/"],
    // Yarn PnP: keep the package unpacked on disk so the binary stays a real,
    // executable file rather than being kept inside a zip.
    preferUnplugged: true,
    publishConfig: { access: "public" },
  });

  writeFileSync(
    join(pkgDir, "README.md"),
    `# ${name}\n\nThe \`${t.os}-${t.cpu}\` prebuilt binary for **nb** (notebook-cli).\n\n` +
      `This is an internal platform package. Install the main package instead:\n\n` +
      "```sh\nnpm install -g @psielta/notebook-cli\n```\n\n" +
      `Docs: https://github.com/psielta/notebook-cli\n`
  );

  optionalDependencies[name] = version; // exact-version lockstep with the main package
}

// --- Main package ------------------------------------------------------------
// Read the committed main package.json as the base (single source of truth for
// description/keywords/etc.) and patch only the version-dependent fields.
const mainDir = join(outDir, `${SCOPE}/notebook-cli`);
console.log(`  • ${SCOPE}/notebook-cli (launcher)`);

const mainPkg = JSON.parse(readFileSync(join(mainSrcDir, "package.json"), "utf8"));
mainPkg.version = version;
mainPkg.optionalDependencies = optionalDependencies;
mainPkg.publishConfig = { ...(mainPkg.publishConfig || {}), access: "public" };

// Copy the committed launcher (and README) verbatim, then write the patched manifest.
cpSync(join(mainSrcDir, "bin"), join(mainDir, "bin"), { recursive: true });
chmodSync(join(mainDir, "bin", "nb"), 0o755);
if (existsSync(join(mainSrcDir, "README.md"))) {
  cpSync(join(mainSrcDir, "README.md"), join(mainDir, "README.md"));
}
writeJson(join(mainDir, "package.json"), mainPkg);

// --- Summary -----------------------------------------------------------------
console.log(`\nDone. Packages written to: ${outDir}`);
if (process.platform === "win32") {
  console.warn(
    `\nWARNING: built on Windows — the linux/macOS binaries are NOT marked executable,\n` +
      `so this build is for LOCAL TESTING ONLY. Publish the official release from\n` +
      `Linux/macOS (the release CI does); scripts/publish-npm.mjs will refuse the\n` +
      `non-executable unix packages.`
  );
}
console.log(`Next: node scripts/publish-npm.mjs            (add --provenance in CI)`);
