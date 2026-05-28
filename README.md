# Notebook CLI (`nb`)

[![CI](https://github.com/psielta/notebook-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/psielta/notebook-cli/actions/workflows/ci.yml)

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
- `sessions/<id>.current`: notebook selecionado na sessao de terminal
  identificada por `<id>`. Cada janela do PowerShell mantem sua propria
  selecao; abrir uma segunda janela comeca sem notebook selecionado ate
  rodar `nb use <nome>`. O `<id>` e derivado do processo do terminal pai
  (PID + horario de criacao) ou do valor opcional de `NB_SESSION_ID`.

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

Build com versao:

```powershell
go build -ldflags "-X notebook-cli/internal/version.Version=v1.0.0" -o nb.exe ./cmd/nb
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
| `nb --version` | Mostra a versao do binario. |

### Tasks diarias

Use `nb new task` para criar e selecionar automaticamente um notebook de task
com nome padronizado pela data local:

```powershell
nb new task
```

Exemplos de nomes gerados:

```text
T001-2026-05-27
T002-2026-05-27
```

A sequencia e diaria e usa os notebooks ja existentes no banco local.

### Atalhos para agentes de IA

`nb add` aceita atalhos fixos para registrar rapidamente a etapa atual de um
agente:

| Atalho | Nota gravada |
| --- | --- |
| `clau-p` | `Claude planejando...` |
| `clau-r` | `Claude revisando...` |
| `clau-i` | `Claude implementando...` |
| `clau-vp` | `Claude validando plano do Codex...` |
| `clau-cp` | `Claude corrigindo plano com melhorias do Codex...` |
| `dex-p` | `Codex planejando...` |
| `dex-r` | `Codex revisando...` |
| `dex-i` | `Codex implementando...` |
| `dex-vp` | `Codex validando plano do Claude...` |
| `dex-cp` | `Codex corrigindo plano com melhorias do Claude...` |
| `clau-to-dex` | `Claude gerando prompt para Codex...` |
| `dex-to-clau` | `Codex gerando prompt para Claude...` |

Textos livres continuam funcionando normalmente. Apos adicionar qualquer nota,
o comando mostra a nota gravada com ID, horario e texto resolvido.

### Cores

A CLI usa cores ANSI automaticamente em terminal interativo e desliga cores em
redirecionamentos, pipes, testes e quando `NO_COLOR` estiver definido.

## Roadmap

Melhorias planejadas para evoluir o projeto:

- `nb remove last`: remover a ultima nota sem precisar copiar o ID.
- `nb edit <id>`: editar o texto de uma nota existente.
- `nb search <termo>`: buscar notas por palavra ou trecho de texto.
- `nb rename <nome-atual> <novo-nome>`: renomear um notebook.
- `nb export`: exportar notas para Markdown ou JSON.
- `nb import`: importar notas a partir de JSON.
- `nb stats`: mostrar quantidade de notebooks, notas e data da ultima nota.
- Instalador simples para Windows, copiando `nb.exe` para uma pasta no `PATH`.
- Releases no GitHub com binarios prontos para download.

## Arquitetura

O projeto usa layout idiomatico com `cmd/` e `internal/`:

```text
cmd/nb                 entrypoint da CLI
internal/commands      comandos Cobra
internal/service       regras de negocio
internal/repository    persistencia com GORM
internal/domain        entidades Notebook e Note
internal/database      abertura e migracao do SQLite
internal/session       resolucao do identificador da sessao de terminal
internal/state         leitura/escrita do current por sessao
internal/version       versao do binario para --version e builds de release
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

- Total: 87.2%
- `internal/repository`: 93.0%
- `internal/service`: 95.6%

## Observacoes Para Windows

PowerShell 7 ja usa UTF-8 por padrao. Em consoles antigos, rode:

```powershell
chcp 65001
```

### Compartilhando a selecao entre sessoes

Por padrao, cada janela do PowerShell tem sua propria selecao de notebook
atual. Para forcar duas janelas a compartilharem a selecao (util para
automacao, CI ou pipes), defina `NB_SESSION_ID` em cada uma:

```powershell
$env:NB_SESSION_ID = "shared"
nb use erp
```

A variavel e escopada ao processo do PowerShell; nao use `setx`, pois ele
grava no registro do usuario e nao afeta sessoes ja abertas. Use um valor
curto com letras, numeros, `_`, `-` ou `.`, mas nao comece com `sh-`, que e
reservado para sessoes detectadas automaticamente.

## Licenca

Distribuido sob a licenca MIT. Veja [LICENSE](LICENSE).
