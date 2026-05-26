# Notebook CLI (`nb`)

CLI em Go para criar e gerenciar notebooks de notas locais diretamente pelo terminal.

O projeto foi pensado para uso no PowerShell do Windows, com armazenamento local em SQLite, comandos curtos e uma arquitetura simples de manter.

## Visao Geral

`nb` organiza notas em notebooks independentes. Cada notebook pode ser selecionado como atual, e as notas podem ser adicionadas, listadas, removidas ou limpas sem sair do terminal.

Os dados ficam no perfil do usuario:

```powershell
$env:USERPROFILE\.notebook-cli
```

Arquivos criados:

- `notebook.db`: banco SQLite unico da aplicacao.
- `.current`: notebook selecionado no momento.

## Funcionalidades

- Criacao e selecao de notebooks.
- Notas com ID local por notebook.
- IDs monotonicos: uma nota removida nao tem seu ID reutilizado.
- Listagem completa ou das ultimas notas.
- Limpeza com confirmacao interativa ou flag `--yes`.
- Armazenamento local em SQLite sem dependencia de CGO.
- Testes unitarios, integracao, comandos Cobra e E2E opcional.

## Instalacao

Requisitos:

- Go 1.22 ou superior.
- PowerShell 7 recomendado no Windows.

Build local:

```powershell
go build -o nb.exe ./cmd/nb
```

Depois, use `.\nb.exe` no diretorio do projeto ou adicione o binario ao `PATH`.

## Uso Rapido

```powershell
.\nb.exe new erp
.\nb.exe use erp
.\nb.exe add "corrigir problema x"
.\nb.exe add "testar importacao de XML"
.\nb.exe show
```

Saida esperada:

```text
Notebook 'erp' criado.
Usando 'erp'.
Nota 1 adicionada.
Nota 2 adicionada.
ID  CRIADO EM            TEXTO
1   2026-05-26 08:45:45  corrigir problema x
2   2026-05-26 08:45:45  testar importacao de XML
```

## Comandos

| Comando | Descricao |
| --- | --- |
| `nb new <nome>` | Cria um novo notebook. |
| `nb use <nome>` | Define o notebook atual. |
| `nb current` | Mostra o notebook atual. |
| `nb list` | Lista notebooks e quantidade de notas. |
| `nb add "<texto>"` | Adiciona uma nota ao notebook atual. |
| `nb show` | Lista todas as notas do notebook atual. |
| `nb last [n]` | Mostra as ultimas `n` notas. Sem argumento, mostra a ultima. |
| `nb remove <id>` | Remove uma nota pelo ID local. |
| `nb clear [--yes]` | Remove todas as notas do notebook atual. |

## Arquitetura

O projeto usa layout idiomatico com `cmd/` e `internal/`:

```text
cmd/nb                 entrypoint da CLI
internal/commands      comandos Cobra
internal/service       regras de negocio
internal/repository    persistencia com GORM
internal/domain        entidades Notebook e Note
internal/database      abertura e migracao do SQLite
internal/state         leitura/escrita do .current
internal/output        formatacao de tabelas no terminal
internal/testutil      helpers para testes
```

Decisoes tecnicas principais:

- `github.com/spf13/cobra` para parsing dos comandos.
- `gorm.io/gorm` com `github.com/glebarez/sqlite`, evitando CGO no Windows.
- SQLite unico em `%USERPROFILE%\.notebook-cli\notebook.db`.
- DSN com `_txlock=immediate`, `foreign_keys(1)`, `journal_mode(WAL)` e `busy_timeout(5000)`.
- Criacao de nota em transacao unica: incrementa `NextNoteID` e insere a nota no mesmo bloco.
- IO dos comandos via Cobra (`SetIn`, `SetOut`, `SetErr`) para facilitar testes.

## Testes

```powershell
go test ./...
go test -tags=e2e ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Cobertura atual:

- Total: 88.1%
- `internal/repository`: 94.4%
- `internal/service`: 92.3%

## Observacoes Para Windows

PowerShell 7 ja usa UTF-8 por padrao. Em consoles antigos, rode:

```powershell
chcp 65001
```

## Licenca

Distribuido sob a licenca MIT. Veja [LICENSE](LICENSE).
