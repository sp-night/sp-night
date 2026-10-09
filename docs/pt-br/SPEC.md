# A especificação do SP Night
<!-- source: 5267621eaed4 -->

Este documento existe para responder a uma pergunta: **quando eu levo o tema para uma nova plataforma, qual cor vai onde?** Sem uma resposta escrita, cada port vira uma interpretação diferente do tema, e foi assim que os ports do Dracula foram se afastando uns dos outros.

As mesmas regras estão publicadas em [sp-night.github.io/spec](https://sp-night.github.io/spec). Esta cópia é a que a ferramenta aplica.

<a id="the-three-layers"></a>

## As três camadas

```
palette/sp_night.json     23 colours per flavour. São Paulo names.
        ↓
palette/roles.json        semantic roles → a palette key
        ↓
<app>.tmpl                asks for a role, never the raw palette
```

**Regra rígida: um mapeamento nunca referencia uma cor da paleta diretamente.**

```
✗  palette = 4={{ .C.marginal }}
✓  palette = 4={{ .R.ansi.blue }}
```

Por quê: mover `syntax.keyword` de `marginal` para `temporal` tem que repintar Neovim, bat, fish e o site de uma vez. Um mapeamento com `marginal` escrito nele fica para trás e ninguém percebe.

O `spn lint` falha na primeira forma. Ele percorre a árvore de parse, não o texto, então um contraexemplo dentro de um comentário não conta como achado.

A exceção são as listas de variáveis (`waybar.css`, `gtk.css`, `hyprland.conf`), que publicam a paleta crua como `@define-color` / `$var` porque o usuário final vai querer escrever `@sp_sodio` na própria folha de estilo. Elas declaram `raw_palette: true` no frontmatter. Mesmo nelas, tudo o que o próprio template estiliza continua usando um papel.

<a id="the-palette"></a>

## A paleta

As quatro faixas abaixo não são títulos deste documento: são o bloco `groups` de `sp_night.json`, e elas **particionam** a paleta. Toda cor pertence a exatamente uma faixa, e a ferramenta recusa uma paleta em que alguma esteja faltando, duplicada ou desconhecida.

Elas são dados, e não Go, porque o site incorpora o contrato e o publica. Quando a partição vivia na ferramenta, o site tinha que redigitá-la, e uma cor que a paleta declarava mas a cópia do site não conhecia sumia da paleta publicada sem nada falhar em lugar nenhum. `accents` e `vivo` são as duas faixas que a auditoria chama pelo nome; o resto de uma faixa é o rótulo e os membros.

<a id="surfaces--names-neutral-about-lightness"></a>

### Superfícies: nomes neutros quanto à luminosidade

Esses nomes descrevem **função** e profundidade, não um hex específico. É isso que permite que um único mapeamento sirva a todas as variações.

| chave | função | referência |
|---|---|---|
| `vao` | o recuo mais fundo: janela flutuante, popup, barra de abas | o vão livre do MASP |
| `laje` | **o fundo padrão**, o plano sobre o qual tudo fica | uma laje de concreto |
| `concreto` | painéis, cards, cursorline, statusline | concreto aparente |
| `vidro` | seleção, modo visual, item ativo | vidro refletindo a rua |
| `fiacao` | bordas, divisórias, guias de indentação | a fiação aérea |

<a id="text"></a>

### Texto

| chave | uso |
|---|---|
| `fg_vivo` | texto padrão em negrito, `ansi.bright_white` |
| `fg` | texto principal, identificadores, variáveis |
| `fg_dim` | comentários, pontuação, texto secundário |
| `fg_muted` | números de linha, guias, texto desabilitado |

<a id="accents"></a>

### Acentos

| chave | ANSI | referência |
|---|---|---|
| `brasa` | red | o MASP, luzes de freio |
| `sodio` | — | a lâmpada de sódio dos postes |
| `taxi` | yellow | semáforo, táxi |
| `ibira` | green | Ibirapuera, sinal verde |
| `estaiada` | — | a ponte Estaiada iluminada |
| `sereno` | cyan | o sereno antes do amanhecer |
| `marginal` | blue | a Marginal, o metrô |
| `temporal` | magenta | o céu logo antes da chuva |

<a id="bright-ansi"></a>

### ANSI brilhante

Os seis brilhantes do terminal (`brasa_vivo`, `taxi_vivo`, `ibira_vivo`, `sereno_vivo`, `marginal_vivo`, `temporal_vivo`) são a mesma cor com +0.06 de luminosidade Oklch, com matiz e croma intactos. Texto em negrito no terminal usa o brilhante; se o brilhante for igual ao normal, essa informação desaparece. Eles ficam na paleta em vez de serem derivados num template, então passam pela mesma auditoria de contraste que todo o resto.

`fg_vivo` é esse mesmo aumento aplicado à rampa de texto, e existe pelo mesmo motivo. Um alvo que separa o texto padrão do texto padrão em **negrito** (o `colors.primary.bright_foreground` do Alacritty, a cor da fonte em negrito do kitty) não tinha nada acima de `ui.fg` para onde apontar, então o mapeamento precisava escrever `fg` nas duas chaves e alegar um aumento que não existia. `ui.fg_bright` dá nome a ele, e `ansi.bright_white` resolve para ele. `ansi.white` continua `fg_dim`: a tabela ANSI fica com as duas pontas da rampa, e `fg` continua sendo o `foreground` simples do terminal.

Para a rampa de texto, +0.06 é o teto, não uma constante herdada. Em `noite`, `fg` já está em Oklch L 0.88; com +0.07 o canal azul satura e o croma começa a cair (0.028 → 0.023). Com +0.06 as três variações recebem o aumento exato com o croma intacto. Foi medido, e foi isso que decidiu o valor.

`sodio` é a cor assinatura: cursor, borda ativa, workspace ativo, relógio, título. Se um port precisa da "cor do tema", é ela, **num terminal ou numa barra**. Num widget de aplicação (seleção e foco em GTK/Qt/KDE) o acento é `ui.accent_alt`, azul: um acento de sistema laranja faz um gerenciador de arquivos inteiro parecer um aviso. O Tokyo Dark faz a mesma divisão, e a cor primária dele sempre foi azul.

`sodio` e `estaiada` não têm posição ANSI, porque as 16 posições já têm dono. Elas existem para a interface, não para o terminal.

<a id="assignment-rules"></a>

## Regras de atribuição

**Sintaxe**: a lógica é que *o que o código faz* fica com azul-ciano, *o que o código é* fica com roxo-laranja, e *dados literais* ficam com verde-laranja.

| papel | cor |
|---|---|
| keyword, conditional, repeat | `marginal` |
| function, method, operator | `sereno` |
| type, namespace | `temporal` |
| constant, number, boolean | `sodio` |
| string, character | `ibira` |
| parameter, macro, escape | `estaiada` |
| builtin, attribute | `taxi` |
| tag | `brasa` |
| variable, property, field | `fg` |
| comment, punctuation | `fg_dim` |

As variáveis ficam em `fg` de propósito. Um tema em que tudo é colorido não tem hierarquia: o olho precisa de descanso, e o identificador comum é o maior volume de texto na tela.

**Diagnósticos e git** seguem a convenção universal e não devem ser reinterpretados: erro = vermelho, aviso = amarelo, info = azul, dica = ciano, ok/adicionado = verde, modificado = amarelo, removido = vermelho.

**Um diff tem um único padrão**, em todos os alvos: adicionado = `ibira`, modificado e renomeado = `taxi`, removido e conflito = `brasa`. Toda a estrutura (hunk, cabeçalho, índice, não rastreado) é neutra (`fg_dim` / `fg_muted`). Um diff que também usa roxo, azul e laranja vira um arco-íris onde só três estados importam.

<a id="surface-temperature"></a>

## Temperatura da superfície

A oposição azul-amarelo é o eixo mais forte que um tema tem, e num tema escuro ela **não** se resolve nos acentos: o azul tem luminância intrinsecamente baixa, então um azul que passa no AA sobre `concreto` precisa ser claro e dessaturado, o oposto de um polo. Medido: repintar os acentos move o eixo de 1.03 para 1.35 e custa o caráter da paleta inteira.

Ela se resolve no **fundo**, que não é texto e por isso não tem restrição de contraste. `noite` tem uma `laje` azul-violeta com croma 0.037; os acentos continuam quentes. A tensão entre o campo e a massa de acentos vai de 0.020 para 0.051 sem mudar um único acento.

Daí a divisão entre as duas variações escuras: `noite` é fria e **saturada**, `garoa` é cinza **chapado** (croma 0.005). A garoa não esfria a cidade, ela a desbota; a diferença é saturação, não temperatura.

<a id="the-lesson-of-the-robust-themes-kde"></a>

## A lição dos temas robustos (KDE)

O que faz o Breeze Dark e o Catppuccin funcionarem em qualquer app não é a paleta, é a disciplina do esquema:

1. **Um único conjunto de cores de primeiro plano, repetido literalmente** em View, Window, Button, Tooltip, Complementary e Header. Nada é derivado por seção; consistência é a ausência de derivação.
2. **A seção Selection é autocontida.** Cada cor de primeiro plano nela é escolhida para ser lida *sobre o acento*, e o `BackgroundAlternate` dela é vizinho do acento, nunca o alternate geral. Foi essa substituição que produziu 1.11:1 no Dolphin.
3. **`[WM]` é escuro e calmo.** Uma barra de título nunca recebe o acento.

E uma que nenhum deles tem: a ferramenta analisa o arquivo `.colors` que acabou de gerar e falha o build se qualquer par fg/bg de qualquer seção ficar abaixo de 4.5:1 (`internal/audit/kde.go`). O Breeze de verdade tem pares de 2–3:1 em Selection; o SP Night não compila com nenhum.

<a id="contrast"></a>

## Contraste

O `spn check` mede cada par texto/superfície de cada variação e **falha o build** quando algum não atinge o mínimo. O piso depende do que é a superfície:

| superfície | piso | por quê |
|---|---|---|
| `laje`, `vao`, `concreto` | 4.5:1 (AA) | onde você lê código por horas |
| `vidro` | 3.0:1 | seleção: um estado passageiro, você olha a forma |
| `fg_muted` sobre qualquer uma | 3.0:1, aviso | ornamento, não texto de leitura |
| `fiacao` sobre `laje`/`vao` | 1.5:1, aviso | uma borda, não texto |

74 pares por variação. Uma verificação que falha interrompe o build; um aviso é reportado.

`fg_dim` é o caso delicado: comentários precisam de AA completo sobre `laje`. Comentário ilegível é o problema número um de todo tema escuro popular, e sempre porque ninguém mediu.

<a id="separation-between-accents"></a>

### Separação entre acentos

Contraste contra o fundo não basta. Dois acentos podem passar no AA com folga e ainda assim ser confundidos **entre si**; foi assim que `estaiada` e `sereno` escaparam da primeira versão desta auditoria: luminância quase idêntica sobre `laje`, distinguíveis só pelo matiz.

ΔE sozinho não é um limiar utilizável. Numa paleta de oito acentos cobrindo o círculo de matizes, matizes vizinhos caem naturalmente em ΔE 0.07–0.09 no Oklab; exigir mais de cada par seria exigir uma paleta com menos cores. Medido nas variações, os pares realmente confundíveis não são os de menor ΔE, e sim os de menor **ΔL**.

A regra: um par perceptualmente vizinho (ΔE < 0.10) precisa se separar pela luminosidade (ΔL ≥ 0.04). É isso que mantém os dois distinguíveis em escala de cinza e, pelo mesmo motivo, para quem tem deficiência na visão de cores.

<a id="colour-vision-deficiency"></a>

### Deficiência na visão de cores

O `spn check -v` também simula protanopia, deuteranopia e tritanopia (matrizes de Viénot 1999) e informa quantos pares de acentos ficam próximos demais em cada uma.

Isso é um **diagnóstico, não uma barreira**, por um motivo concreto: separar oito acentos sob daltonismo, manter o caráter enfumaçado da paleta e passar no AA sobre um fundo escuro é um sistema sobredeterminado. Uma busca em matiz, croma e luminosidade encontra soluções com separação muito melhor, mas todas chegam lá levando o tema para outro lugar: o amarelo vira oliva, o vermelho vira terracota, os cianos batem na borda do gamut e ficam neon. Nenhum tema popular resolve as três coisas; Catppuccin, Tokyo Night e Dracula também não.

O número existe para que a escolha seja deliberada. Se a prioridade mudar algum dia, a ferramenta já mede isso.

<a id="about-the-δl-rule"></a>

### Sobre a regra do ΔL

Ela tem um ponto cego que vale conhecer: só examina pares de matiz vizinho na visão normal. Na deuteranopia, os pares que colapsam são os *distantes*: vermelho e verde caem no mesmo lugar. Uma medida não cobre a outra, e por isso as duas coexistem.

É um aviso, não uma falha: essas são as cores de identidade do tema, e a decisão de repintá-las é de quem mantém o projeto. A escada de luminosidade dos acentos mantém a lista vazia hoje.

Rode `spn check` antes de mexer em uma cor.

<a id="adding-a-colour"></a>

## Adicionar uma cor

1. Adicione-a em `colors` em **todas** as variações, e em `meaning`.
2. Adicione-a à faixa a que pertence em `groups`.
3. `spn check`, depois `go test ./...`.

O passo 2 não é burocracia opcional: uma cor fora de todas as faixas falha na validação, porque tudo o que percorre a paleta percorre as faixas. Se for texto ou acento, decida o piso que ela precisa atingir (`internal/audit` nomeia as chaves que mede) e, se for um `_vivo`, ela aumenta a base em 0.06 de luminosidade Oklch, e a suíte verifica que isso acontece.

<a id="adding-a-flavour"></a>

## Adicionar uma variação

1. Adicione o bloco em `palette/sp_night.json` → `flavors`.
2. `spn check`.

Só isso: a ordem vem do próprio JSON, então não há uma segunda lista em Go para manter em sincronia.

Uma nova variação precisa se distinguir das existentes pelo **matiz**, não só pela luminosidade. `garoa` e `noite` são o exemplo: a primeira versão da garoa era a noite com os pretos clareados, e dos ΔE 0.034 entre as superfícies, 0.033 eram ΔL, a mesma cor mais clara. Puxar o cinza para o azul (Oklab B de −0.005 para −0.017) aumentou o Δcroma 7× e reduziu o ΔL: elas viraram duas atmosferas em vez de dois brilhos.

Nenhum mapeamento muda. Se um mapeamento precisa mudar para acomodar uma nova variação, é sinal de que ele está usando a paleta crua onde deveria usar um papel.

<a id="adding-a-target"></a>

## Adicionar um alvo

1. Liste-o em `registry/ports.yml`.
2. `spn new <slug>`.
3. Escreva o mapeamento usando `.R`.

O frontmatter aceita um `matrix` vazio para renderizar uma vez só, no caso de um app que lê um único arquivo de configuração fixo. Formatos que precisam de um bloco claro ao lado do escuro espelham o lado escuro, já que o tema é só escuro; `.All` alcança as outras variações: `{{ (index .All "garoa").R.ui.bg }}`.

Para "texto legível sobre este acento", use `readable`, que escolhe entre duas opções pelo contraste medido em vez de chutar:

```
"mOnPrimary": "{{ readable .C.vao .C.fg .R.ui.accent }}"
```

Helpers disponíveis num mapeamento: `nohash`, `rgb`, `rgbn`, `rgba`, `hexa`, `argb`, `sgrfg`, `sgrbg`, `mix`, `lighten`, `darken`, `contrast`, `readable`, `kebab`, `pad`, `repeat`, `tojson`, `upper`, `lower`, `r`, `g`, `b`.
