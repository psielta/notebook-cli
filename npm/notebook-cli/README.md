# nb — notebook-cli

[![npm](https://img.shields.io/npm/v/%40psielta%2Fnotebook-cli.svg)](https://www.npmjs.com/package/@psielta/notebook-cli)

A fast, local-first **notes CLI**. Organize notes into notebooks straight from your
terminal. Cross-platform, single local SQLite store, no daemon, no config.

> `nb` is a small **Go** binary. This npm package ships prebuilt binaries for each
> platform as optional dependencies, so installing is instant — there is no
> postinstall step and no download at install time.

## Install

```sh
npm install -g @psielta/notebook-cli
```

Or run it once without installing:

```sh
npx @psielta/notebook-cli new erp
```

Supported platforms: **Linux**, **macOS** and **Windows** on **x64** and **arm64**.

## Usage

```sh
nb new erp                 # create a notebook
nb use erp                 # select it as current
nb add "fix bug x"         # add a note
nb add "test xml import"
nb show                    # list notes in the current notebook
nb last 5                  # show the last 5 notes
nb list                    # list notebooks
nb --version
```

Data is stored locally under your home directory (`~/.notebook-cli/notebook.db`).
The current notebook selection is scoped per terminal session.

## Documentation

Full docs, command reference and architecture: <https://github.com/psielta/notebook-cli>

## License

MIT
