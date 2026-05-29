# Releasing & npm publishing

`nb` is a Go binary distributed on npm using the **optionalDependencies** pattern
(the same approach as esbuild and Biome): the public package
[`@psielta/notebook-cli`](https://www.npmjs.com/package/@psielta/notebook-cli)
is a tiny Node launcher, and the actual binary ships as a per-platform optional
dependency (`@psielta/nb-<os>-<arch>`). npm installs only the one matching the
user's `os`/`cpu`, so there is **no postinstall and no download** at install time.

```
@psielta/notebook-cli          main package: bin/nb (Node launcher) + optionalDependencies
@psielta/nb-linux-x64          prebuilt binary, os:[linux]  cpu:[x64]
@psielta/nb-linux-arm64                          os:[linux]  cpu:[arm64]
@psielta/nb-darwin-x64                           os:[darwin] cpu:[x64]
@psielta/nb-darwin-arm64                         os:[darwin] cpu:[arm64]
@psielta/nb-win32-x64                            os:[win32]  cpu:[x64]
@psielta/nb-win32-arm64                          os:[win32]  cpu:[arm64]
```

Because `nb` is built with `CGO_ENABLED=0`, each Linux binary is fully static and
runs on both glibc and musl (Alpine) — no separate musl packages are needed.

## One-time setup

### 1. npm account / scope

The packages are published under the **`@psielta`** scope, so your npm account
must own that scope (username `psielta`, or an org named `psielta`). If your npm
name differs, change the scope in **three** places:

- `SCOPE` in [`npm/notebook-cli/bin/nb`](npm/notebook-cli/bin/nb)
- `SCOPE` in [`scripts/build-npm.mjs`](scripts/build-npm.mjs)
- `name` (and it will regenerate `optionalDependencies`) in [`npm/notebook-cli/package.json`](npm/notebook-cli/package.json)

If you fork or transfer the **GitHub repo** to a different owner/name, also update
`release.github.owner`/`name` in [`.goreleaser.yaml`](.goreleaser.yaml) and the
`repository`/`homepage`/`bugs` URLs in the package manifests (provenance requires
the `repository` URL to match the repo running the Action — see below).

> The committed [`npm/notebook-cli/package.json`](npm/notebook-cli/package.json) is
> a **template**: its `version` and `optionalDependencies` (placeholder `0.0.0`)
> are rewritten by `build-npm.mjs` at build time. Don't publish it directly —
> always go through the scripts. `publish-npm.mjs` refuses `0.0.0`/`-dev` versions.

### 2. Create the NPM_TOKEN secret

Classic automation tokens were disabled by npm in November 2025 — use a
**Granular Access Token**:

1. npmjs.com → avatar → **Access Tokens → Generate New Token → Granular Access Token**.
2. Expiration: ≤ 90 days (npm's max for write tokens — set a reminder to rotate).
3. Permissions: **Packages and scopes → Read and write**, scoped to **`@psielta`**.
4. Enable the token's **Bypass 2FA** option (write tokens enforce 2FA by default,
   which would block non-interactive CI), **or** do the very first publish locally
   with interactive 2FA and tighten per-package settings afterward.
5. Copy the `npm_…` value and add it as a repo secret:
   **Settings → Secrets and variables → Actions → New repository secret**,
   named exactly **`NPM_TOKEN`**.

> The first publish is a chicken-and-egg: the `@psielta/*` packages don't exist
> yet, so per-package 2FA settings can't be set in advance. Use the token's
> Bypass-2FA option (or a local first publish) to bootstrap, then tighten.

After the first successful publish, tighten security: on npmjs.com → each
`@psielta/*` package → **Settings → Require 2FA**, choose
*"Require two-factor authentication or an automation/granular token"* (not
*"...and disallow tokens"*, which would block CI).

## Cutting a release

Everything is automated by [`.github/workflows/release.yml`](.github/workflows/release.yml).
Just push a semver tag:

```sh
git tag v1.2.3
git push origin v1.2.3
```

The workflow then, on `ubuntu-latest` (Linux keeps the binaries' exec bit in the
npm tarballs):

1. runs `go test ./...`;
2. runs **GoReleaser** → builds binaries, creates the GitHub Release with
   `.tar.gz`/`.zip` archives + checksums + changelog;
3. runs `scripts/build-npm.mjs` → cross-compiles all six binaries and assembles
   the npm packages into `dist-npm/`;
4. runs `scripts/publish-npm.mjs --provenance` → publishes the platform packages
   first, then the main package, each with **build provenance** and
   `--access public`. Re-running is safe: versions already on the registry are
   skipped.

The version (`v1.2.3`) is injected into the binary via
`-ldflags "-X notebook-cli/internal/version.Version=v1.2.3"`, so `nb --version`
reports it; the npm version is the same string without the leading `v` (`1.2.3`).

### Re-running a failed release

The npm side is **idempotent** — `publish-npm.mjs` skips any package version
already on the registry, so re-running finishes the rest. **GoReleaser is not**:
if the GitHub Release for the tag already exists it errors with "release already
exists". To re-run a release that failed partway:

```sh
gh release delete v1.2.3 --yes   # delete the partial GitHub Release (keeps the tag)
```

then re-run the workflow (Actions → Release → Re-run, or push the tag again). The
already-published npm packages are skipped automatically. The workflow's
`concurrency` group also prevents two release runs for the same tag overlapping.

## Testing the npm packaging locally

> The commands below assume a POSIX shell (bash/zsh). On Windows run them in Git
> Bash or WSL — or just rely on the CI, which does this on Linux.

You don't need to publish to validate the packaging. On any platform with Go +
Node:

```sh
node scripts/build-npm.mjs --version 0.0.0-test
node scripts/publish-npm.mjs --dry-run        # shows the publish plan, publishes nothing

# Verify exactly the binary is packed (no stray files):
npm pack --dry-run ./dist-npm/@psielta/nb-$(node -p process.platform)-$(node -p process.arch)

# Smoke-test the launcher against your host binary:
NB_BINARY="$(pwd)/dist-npm/@psielta/nb-$(node -p process.platform)-$(node -p process.arch)/bin/nb" \
  node ./npm/notebook-cli/bin/nb --version
```

> Note: a binary cross-built **on Windows** for Linux/macOS won't carry the unix
> exec bit, so publish the official release from CI (Linux) — never from a Windows
> laptop.

## Provenance requirements (already wired up)

For `npm publish --provenance` to succeed, all of the following must hold — the
workflow and generated manifests already satisfy them:

- `permissions: id-token: write` on the job (mints the OIDC token);
- every `package.json` has a `repository` URL pointing at
  `github.com/psielta/notebook-cli` (the repo running the Action);
- packages are public (`publishConfig.access: "public"` + `--access public`);
- publishing happens from CI (provenance can't be generated from a laptop);
- npm ≥ 9.5.0 (Node 22's npm 10.x is fine).

## Optional: tokenless publishing (OIDC Trusted Publishing)

Long-term you can drop `NPM_TOKEN` entirely by configuring a **Trusted Publisher**
per package (npmjs.com → package → Settings → Trusted Publisher → GitHub Actions,
repo `psielta/notebook-cli`, workflow `release.yml`). Then remove
`--provenance` (it becomes automatic) and `NODE_AUTH_TOKEN`, keep
`--access public` and `id-token: write`, and use npm ≥ 11.5.1. Bootstrap the first
version with the token path, then switch.
