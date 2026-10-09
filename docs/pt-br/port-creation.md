# Criar um port
<!-- source: 6ff717326d69 -->

Um port é um **mapeamento**, não uma cópia de cores: ele decide qual chave da configuração de um app recebe qual papel, e a ferramenta resolve o resto. Leia primeiro o [SPEC.md](SPEC.md), em especial a regra de que um mapeamento nunca referencia a paleta crua.

O exemplo canônico é [`sp-night/ghostty`](https://github.com/sp-night/ghostty).

<a id="non-negotiables"></a>

## Inegociáveis

1. **Nenhum hex é escolhido à mão.** Nem num arquivo de tema, nem num README, nem numa prévia. Se algo precisa mudar, muda o mapeamento ou o catálogo, e tudo é regenerado.
2. **Conventional Commits**, em inglês, em minúsculas, sem ponto final: `feat(themes):`, `docs(readme):`, `feat(ports):`, `ci:`, `test:`, `chore:`.
3. **Nunca faça commit direto na `main`.** Branch → pull request → merge.
4. **Documentação pública em inglês.** Os nomes das cores e das variações ficam em português: eles são a identidade.
5. **MIT** em todos os repositórios.

<a id="1--list-the-target-in-the-catalogue"></a>

## 1 — Listar o alvo no catálogo

`registry/ports.yml` é de onde o README, a prévia e o cabeçalho gerado leem. Nada pode ser gerado para um port que não está listado.

Comece de uma entrada impressa em vez de uma página em branco:

```sh
spn entry <slug> >> registry/ports.yml
```

Isso escreve todos os campos obrigatórios, uma tabela de mapeamento para completar e uma prévia completa: a sessão fictícia, trecho por trecho, no formato que as prévias publicadas compartilham. Ela já é válida do jeito que é impressa, então a entrada pode ser commitada enquanto os TODOs ainda estão sendo substituídos. A prévia é só um ponto de partida: reescreva a sessão quando o port tiver uma voz própria, como aconteceu com a do eza.

Os campos que ela deixa para você:

```yaml
  - slug: kitty                 # also the repository name under the org
    name: kitty
    group: terminal
    blurb: One line on what the port covers.
    homepage: https://sw.kovidgoyal.net/kitty/
    repo: https://github.com/sp-night/kitty     # derived from slug, and checked
    install: ~/.config/kitty/sp_night_{flavor}.conf
    activate: include sp_night_{flavor}.conf
    template: kitty.conf.tmpl
    mapping:
      - key: "`background`"
        role: "`ui.bg`"
        meaning: "*laje* under the main text"
    preview:
      title: kitty — sp_night_{flavor}
      frame: terminal             # terminal | editor | app | pane
      swatches: {label: palette 0–15, roles: [ansi.black, ...]}
      body:
        - - {t: "~/sp-night ", r: ui.accent}
          - {t: "❯ ", r: ui.accent_alt}
          - {t: "kitty +kitten themes --dump-theme", r: ui.fg}
```

Depois verifique:

```sh
spn registry --registry registry/ports.yml --copy registry/copy.yml
```

`mapping` não é enfeite: é a tabela que o leitor confere antes de confiar num port, e a validação recusa uma entrada sem ela.

`frame` é a moldura em que a prévia é desenhada, e deve ser o tipo de janela em que o app realmente vive. O body é só a sessão; a moldura desenha o resto, então um port nunca precisa simular uma gutter ou uma statusline trecho por trecho.

| frame | para | o que o renderizador acrescenta | aceita |
|---|---|---|---|
| `terminal` (padrão) | um emulador de terminal | barra de título com três pontos e o `title` centralizado | — |
| `editor` | um editor | uma bufferline com o `title` como aba aberta, linhas numeradas, uma statusline | `bar`, `cursor_line` |
| `app` | uma TUI de tela cheia | `title` como cabeçalho, uma barra inferior | `bar` |
| `pane` | uma CLI que roda no terminal de outra pessoa | nada além da borda | — |

`bar` é `{left: [spans], right: [spans]}`, no mesmo formato de trechos do body; `cursor_line` é a linha do body (contando a partir de 1) que o editor destaca. Cada moldura comporta um número fixo de linhas do body acima da faixa de cores (13 para um pane, 11 para os demais), e o `spn registry` recusa uma sessão que passe disso.

<a id="2--create-the-repository"></a>

## 2 — Criar o repositório

```sh
gh repo create sp-night/<slug> --public --license mit \
  --description "🌃 SP Night for <App> — a dark colour scheme with São Paulo as its reference"
```

O commit inicial traz só o `LICENSE`. Clone-o ao lado dos outros e crie a estrutura:

```sh
spn new <slug>
```

Isso escreve o esboço do mapeamento com o cabeçalho canônico já ligado ao catálogo, o workflow de seis linhas, a configuração do Renovate e as configurações de editor.

<a id="3--write-the-mapping"></a>

## 3 — Escrever o mapeamento

Este é o único arquivo que uma pessoa escreve. Peça um papel, nunca uma cor:

```
✗  background {{ .C.laje }}
✓  background {{ .R.ui.bg }}
```

`spn palette --roles` lista todos os papéis e para o que cada um resolve em cada variação. O que cada um *significa* está no [SPEC.md](SPEC.md#assignment-rules).

O frontmatter declara o contrato:

```yaml
---
spn:
  version: "1.2.0"                                   # the exact engine this was generated with
  matrix: [flavor]                                   # one file per flavour
  filename: "themes/sp_night_{{ .Flavor }}.conf"     # where it goes
---
```

Deixe `matrix` de fora para um app que lê um único arquivo de configuração fixo. O eza é esse caso: o mapeamento dele renderiza uma vez só, e cada variação ainda é publicada como um arquivo em `themes/`.

`version` é uma versão **exata**, não um intervalo, e ninguém a edita à mão. O `spn new` escreve a versão do binário que criou a estrutura do port e, a partir daí, a propagação de release do motor a atualiza no mesmo pull request que regenera os arquivos de tema.

Esse pareamento é o ponto. Os arquivos commitados e a versão que os produziu são um único fato, então o CI de um port instala o motor fixado e verifica os arquivos contra ele. Um intervalo (`^1.0` era o original) faz o CI resolver para o que for mais novo, e o pull request de um port passa de verde para vermelho porque *o motor* lançou uma versão, sem nada ter mudado no port. Também deixava inerte a regra do Renovate que deveria propagar as atualizações do motor: todo `1.x` satisfaz `^1.0`, então nunca havia nada para atualizar.

Depois gere todo o resto:

```sh
spn gen && spn readme && spn preview
spn lint && spn gen --check
```

<a id="4--commit-history"></a>

## 4 — Histórico de commits

Na branch `feat/<slug>-port`, espelhando o ghostty:

```
feat(themes): add the noite, garoa and jaragua flavours
docs(readme): add install guide and palette-generated previews
ci: check the mapping and the generated files
```

Depois `gh pr create`, e faça o merge pelo pull request.

Antes de dar push, confirme que o cabeçalho gerado aponta para o repositório deste port e traz o caminho de instalação e a linha de ativação. Se não trouxer, a entrada do catálogo está errada; não edite um arquivo gerado.

<a id="5--the-website-which-is-not-a-step"></a>

## 5 — O site, que não é um passo

Não há nada a fazer aqui, nem nada a editar. O site incorpora este catálogo em vez de manter a própria lista, então fazer o merge do passo 1 na `main` é o que publica o port: o push dispara o `sync-ports.yml`, cujo job `site` copia `ports.yml` e `copy.yml`, redesenha as prévias, regenera os assets derivados do site, roda a suíte dele e abre um pull request. Fazer o merge desse pull request publica `/ports/<slug>`, uma página montada a partir da entrada, com o guia de instalação, a tabela de chave para papel e uma captura de tela por variação.

Listar um port é publicá-lo. Se você se pegar editando `src/data/` à mão, a mudança vai sumir na próxima sincronização sem deixar rastro.

## Checklist

- [ ] Listado em `registry/ports.yml` (comece com `spn entry <slug>`), `spn registry` sem erros
- [ ] `spn lint` sem erros: o mapeamento pede papéis, nunca cores
- [ ] `spn gen --check`, `spn readme --check`, `spn preview --check` todos sem erros
- [ ] Três variações presentes (noite, garoa, jaragua)
- [ ] Cabeçalho gerado aponta para o repositório deste port
- [ ] `.github/workflows/theme.yml` chama o workflow reutilizável
- [ ] Branch `feat/<slug>-port`, Conventional Commits, merge via pull request
- [ ] Catálogo com merge na `main`, e o pull request de sincronização do site com merge feito

<a id="when-a-port-needs-something-the-roles-do-not-offer"></a>

## Quando um port precisa de algo que os papéis não oferecem

Isso é um achado real, não um motivo para recorrer à paleta. Um papel faltando significa que a camada semântica tem um buraco: abra uma issue neste repositório descrevendo a chave que você não conseguiu mapear. Adicionar um papel repinta todos os ports que o querem; escrever uma cor num único mapeamento ajuda exatamente um port e esconde a lacuna.
