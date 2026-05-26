# notebook-cli

CLI em Go para gerenciar notebooks de notas locais no terminal.

O binario principal e `nb`. Por padrao, os dados ficam em:

```powershell
$env:USERPROFILE\.notebook-cli
```

## Build

```powershell
go build -o nb.exe ./cmd/nb
```

## Uso

```powershell
.\nb.exe new erp
.\nb.exe use erp
.\nb.exe add "corrigir problema x"
.\nb.exe add "testar importacao de XML"
.\nb.exe show
.\nb.exe remove 1
.\nb.exe clear --yes
```

## Comandos

- `nb new <nome>` cria um notebook.
- `nb use <nome>` seleciona o notebook atual.
- `nb current` mostra o notebook atual.
- `nb list` lista notebooks e quantidade de notas.
- `nb add "<texto>"` adiciona uma nota.
- `nb show` lista as notas do notebook atual.
- `nb last [n]` mostra as ultimas notas.
- `nb remove <id>` remove uma nota pelo ID local.
- `nb clear [--yes]` remove todas as notas do notebook atual.

## Testes

```powershell
go test ./...
go test -cover ./...
```

PowerShell 7 ja usa UTF-8. Em consoles antigos, `chcp 65001` pode ajudar na exibicao de acentuacao.
