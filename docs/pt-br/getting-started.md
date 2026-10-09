# Primeiros passos
<!-- source: c7ecc988413f -->

Há dois caminhos. A maioria das pessoas quer um tema para um app que já usa, e para isso não precisa de ferramenta nenhuma. O restante quer criar um port ou reajustar a paleta, e é para isso que serve o `spn`.

<a id="use-a-theme"></a>

## Usar um tema

Cada port é um repositório próprio em [github.com/sp-night](https://github.com/sp-night) e traz arquivos de tema prontos, sem etapa de build. Instalar um é sempre os mesmos três passos:

1. **Escolha uma variação.** `noite` é a padrão, `garoa` é a de cinza chapado, `jaragua` puxa o escuro para o verde. As três são escuras.
2. **Copie o arquivo** dessa variação da pasta `themes/` do port para o caminho que o app lê.
3. **Ative** com a linha que o app espera.

No Ghostty, por exemplo:

```sh
mkdir -p ~/.config/ghostty/themes
curl -fsSL -o ~/.config/ghostty/themes/sp_night_noite \
  https://raw.githubusercontent.com/sp-night/ghostty/main/themes/sp_night_noite
echo 'theme = sp_night_noite' >> ~/.config/ghostty/config
```

O README de cada port, e a página dele em [sp-night.github.io/ports](https://sp-night.github.io/ports), traz o caminho de instalação e a linha de ativação exatos, uma prévia por variação e a tabela de qual chave do app recebe qual papel. Essas páginas são geradas do mesmo catálogo que os arquivos de tema, então não têm como divergir do que você instala.

<a id="install-spn"></a>

## Instalar o spn

Você só precisa do motor para criar um port, mudar um mapeamento ou reajustar a paleta.

```sh
go install github.com/sp-night/sp-night/cmd/spn@latest
spn version
```

Binários de release para Linux e macOS (amd64 e arm64) estão na [página de releases](https://github.com/sp-night/sp-night/releases). A paleta e o catálogo de ports vão dentro do binário, então o `spn` funciona em qualquer diretório.

<a id="build-from-source"></a>

## Compilar do código-fonte

```sh
git clone https://github.com/sp-night/sp-night
cd sp-night
go build ./cmd/spn
go test ./...
go run ./cmd/spn check -v --palette palette
```

Uma dependência, para YAML. A suíte de testes renderiza os mapeamentos dos ports já publicados e compara a saída com os arquivos que os usuários instalaram, byte a byte.

<a id="look-at-the-palette"></a>

## Ver a paleta

```sh
spn palette            # every colour, per flavour
spn palette --roles    # what each role resolves to
spn check -v           # the contrast audit, pair by pair
```

<a id="next"></a>

## Próximos passos

- Antes de escrever ou mudar um mapeamento, leia [a especificação](SPEC.md).
- Para adicionar um app, siga [Criar um port](port-creation.md).
- Todos os comandos e opções estão na [referência de comandos](reference/cli.md).
