# SP Night
<!-- source: e29715b11631 -->

**A lâmpada de sódio pinta a cidade inteira desta cor.** SP Night é um esquema de cores escuro que tem São Paulo como referência: a lâmpada de sódio dos postes, o concreto aparente, o vão livre do MASP, a garoa antes da chuva.

Este repositório é o **contrato** e a **ferramenta**. O contrato é uma paleta e uma camada de papéis que diz qual cor significa o quê. A ferramenta, `spn`, transforma o mapeamento de um port em arquivos de tema prontos, mede a paleta contra pisos de contraste e se recusa a gerar quando ela não os atinge.

<a id="where-to-start"></a>

## Por onde começar

- [Primeiros passos](getting-started.md): escolha uma variação, instale um port ou compile o `spn` você mesmo.
- [A especificação](SPEC.md): a paleta, os papéis e a regra de qual cor vai onde. Leia antes de mexer em um mapeamento.
- [Criar um port](port-creation.md): de uma entrada no catálogo a um repositório publicado, passo a passo.
- [Referência de comandos](reference/cli.md): todos os comandos e opções do `spn`.

<a id="the-three-flavours"></a>

## As três variações

As três são escuras, por decisão: três jeitos de olhar a mesma cidade.

| id | nome | ideia |
|---|---|---|
| `noite` | Noite Paulista | a cidade às 3 da manhã; escuro azul-violeta, a lâmpada de sódio acesa e quente por cima |
| `garoa` | Garoa | a mesma janela através da garoa; cinza chapado, croma quase zero |
| `jaragua` | Pico do Jaraguá | a mesma noite vista do ponto mais alto; o escuro puxado para a mata |

## Links

- Site, com a paleta e uma página por port: [sp-night.github.io](https://sp-night.github.io)
- Guias de instalação de todos os ports: [sp-night.github.io/ports](https://sp-night.github.io/ports)
- Código-fonte: [github.com/sp-night/sp-night](https://github.com/sp-night/sp-night)
- Licença: [MIT](../../LICENSE)
