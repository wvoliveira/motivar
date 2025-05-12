# Motivar

Se mantenha motivado com frases incríveis direto do seu terminal.

## Instale

Basta baixar o binário [aqui](https://github.com/wvoliveira/motivar/releases/) e começar a usar.

```
$ motivar --help
```

```
             ._ o o
             \_´-)|_
          ,""       \
        ,"  ## |   ಠ ಠ.
      ," ##   ,-\__    ´.
    ,"       /     ´--._;)
  ,"     ## / Motivar v0.1.0
,"   ##    /

Usage:
  -debug
        Enable debug mode
  -l string
        Choose a language to show quotes [br,us] (default "br")
Subcommand add-phrases:
  -fmt string
        Specify format phrases content [csv,json] (default "csv")
  -language string
        The language of phrases [br,us]
  -url string
        Specify URL to download from
```

Exemplo: 
```
$ motivar
```

```
O liderado será reflexo da sua liderança, então quem espera lealdade, primeiro deve ser leal. Flávio Augusto.
```

Em inglês

```
$ motivar -l us
```

```
Whatever you are, be a good one. Abraham Lincoln
```

Ou por variavel de ambiente

```bash
# bash
export MOTIVAR_LANGUAGE=us
./motivar

Whatever the mind of man can conceive and believe, it can achieve. Napoleon Hill
```

```powershell
# powershell
$env:MOTIVAR_LANGUAGE = 'us'
.\motivar.exe

Life is fragile. We’re not guaranteed a tomorrow so give it everything you’ve got. Tim Cook
```

Adicionando mais frases via URL

```bash
motivar add-phrases -fmt <json|csv> -language <br|us> -url <url do arquivo>
```

## Funções

- Frases em inglês e português
- Atualize a base de dados direto pelo CLI
- Modifique através de variáveis de ambiente ou arquivo de configuração

### Notes

- Inspirado em [motivate](https://github.com/mubaris/motivate)
- Ascii art: <https://c.r74n.com/textart>
