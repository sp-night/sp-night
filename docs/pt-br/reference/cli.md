# Referência de comandos
<!-- source: 6bc56498f253 -->

```
spn <command> [flags]
```

`spn` sem argumentos, `spn help` ou `spn -h` imprime a lista de comandos. As opções podem vir antes ou depois dos argumentos posicionais: `spn new eza --out x` e `spn new --out x eza` são o mesmo comando.

A paleta e o catálogo de ports vêm embutidos no binário. As opções abaixo que recebem um arquivo ou diretório leem essa cópia do disco, e é assim que o contrato é alterado e medido antes de ser publicado.

<a id="commands-that-write-files"></a>

## Comandos que escrevem arquivos

`--check` em `gen`, `readme` e `preview` não escreve nada e falha quando o que está commitado está desatualizado. É isso que o CI de um port executa.

### spn gen

Renderiza o mapeamento de um port em arquivos de tema prontos.

```sh
spn gen [templates...]
```

Sem argumentos, renderiza todos os `*.tmpl` do diretório atual, então um repositório de port roda apenas `spn gen`.

| opção | função |
|---|---|
| `--check` | não escreve; falha se algum arquivo em disco estiver desatualizado |
| `--stdout` | escreve na saída padrão em vez de em arquivos |
| `--app <slug>` | slug do catálogo para este port (padrão: inferido do nome do template) |
| `--registry <file>` | lê o catálogo de ports deste arquivo |
| `--palette <dir>` | lê o contrato deste diretório |
| `-q` | imprime só as falhas |

### spn readme

Renderiza o README canônico de um port a partir do catálogo.

```sh
spn readme [slug]
```

| opção | função |
|---|---|
| `--out <file>` | arquivo a escrever (padrão `README.md`) |
| `--check` | não escreve; falha se o README estiver desatualizado |
| `--stdout` | escreve na saída padrão em vez de em um arquivo |
| `--registry <file>` | lê o catálogo de ports deste arquivo |
| `--copy <file>` | lê o texto compartilhado dos READMEs deste arquivo |
| `--palette <dir>` | lê o contrato deste diretório |

### spn preview

Desenha a prévia sintética de cada variação. As prévias são SVGs desenhados a partir da paleta, então não têm como mostrar uma cor que o usuário não vai receber.

```sh
spn preview [slug]
```

| opção | função |
|---|---|
| `--out <dir>` | diretório onde escrever as prévias (padrão `assets`) |
| `--check` | não escreve; falha se alguma prévia estiver desatualizada |
| `--registry <file>` | lê o catálogo de ports deste arquivo |
| `--copy <file>` | lê o texto compartilhado dos READMEs deste arquivo |
| `--palette <dir>` | lê o contrato deste diretório |

### spn new

Cria a estrutura de um novo repositório de port: o esboço do mapeamento com o cabeçalho canônico, o workflow de seis linhas, a configuração do Renovate e as configurações de editor. O port precisa estar listado no catálogo antes.

```sh
spn new <slug>
```

| opção | função |
|---|---|
| `--out <dir>` | diretório onde criar a estrutura (padrão: o slug) |
| `--force` | sobrescreve arquivos que a estrutura substituiria |
| `--registry <file>` | lê o catálogo de ports deste arquivo |
| `--copy <file>` | lê o texto compartilhado dos READMEs deste arquivo |

### spn pin

Reescreve a versão do motor que um mapeamento declara no frontmatter. A propagação de release roda esse comando no mesmo pull request que regenera os arquivos de um port.

```sh
spn pin <version> [templates...]
```

| opção | função |
|---|---|
| `--check` | informa se a versão fixada já corresponde, sem escrever nada |

<a id="commands-that-check"></a>

## Comandos que verificam

### spn lint

Verifica se os mapeamentos pedem papéis, nunca cores cruas. Percorre a árvore de parse do template, então um contraexemplo dentro de um comentário não conta como achado.

```sh
spn lint [templates...]
```

### spn check

Audita contraste, separação entre acentos e visão de cores em todas as variações. Falha quando um par fica abaixo do seu piso.

| opção | função |
|---|---|
| `-v` | lista também os pares aprovados e o resumo de visão de cores |
| `--json` | imprime, como dados, o que a auditoria mede e a política contra a qual mede |
| `--palette <dir>` | lê o contrato deste diretório |

### spn registry

Lista e valida o catálogo de ports.

| opção | função |
|---|---|
| `--json` | imprime o catálogo como JSON, usado para montar a matriz de propagação |
| `--registry <file>` | lê o catálogo de ports deste arquivo |
| `--copy <file>` | lê o texto compartilhado dos READMEs deste arquivo |

<a id="commands-that-print"></a>

## Comandos que imprimem

### spn entry

Imprime uma entrada inicial de catálogo para um novo port, com todos os campos obrigatórios, uma tabela de mapeamento para completar e uma prévia completa. Ela já é válida do jeito que é impressa.

```sh
spn entry <slug> >> registry/ports.yml
```

### spn palette

Imprime a paleta e a camada de papéis resolvida.

| opção | função |
|---|---|
| `--roles` | imprime a camada de papéis resolvida em vez das cores |
| `--json` | imprime o contrato como JSON em vez de uma tabela |
| `--palette <dir>` | lê o contrato deste diretório |

### spn version

Imprime a versão da ferramenta. Um build a partir do código-fonte imprime `dev`.
