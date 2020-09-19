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
   ,"     ## / Motivar v0.1
 ,"   ##    /

 Usage of motivar:
  -l string
        Choose a language to show quotes [br,us] (default "br")
  -language string
        Choose a language to show quotes [br,us] (default "br")
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

## Funções

- Frases em inglês e português
- Configurável
- Atualize a base de dados direto pelo CLI
- Adicione endpoints para aumentar a base de frases
- Modifique através de variáveis de ambiente ou arquivo de configuração

### Notes

- Inspirado em [motivate](https://github.com/mubaris/motivate)
- Ascii art: <https://c.r74n.com/textart>
