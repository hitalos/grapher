# Grapher

https://github.com/user-attachments/assets/a30b7935-9312-4503-838c-4c3e6b849625

## Objetivo

Esta aplicação gera um gráfico de barras interativo a partir de uma consulta SQL em uma série temporal de dados. O backend é um serviço HTTP em Go que serve uma página web com o gráfico renderizado em SVG usando a biblioteca D3.js.

## Tecnologias

* **Backend:** Go
* **Frontend:** D3.js, HTML, CSS
* **Banco de Dados:** PostgreSQL

## Features

* **Gráfico Interativo:** Exibe um gráfico de barras com animações e tooltips informativos.
* **Consulta Configurável:** A consulta SQL para buscar os dados é configurável via variáveis de ambiente.
* **Filtragem por Período:** Permite filtrar os dados por um período de tempo através de parâmetros na URL.
* **Executável autosuficiente** O executável já tem embutido os arquivos estáticos.
* **Modo de Desenvolvimento:** Inclui um modo de desenvolvimento que facilita o debug e a visualização de novas alterações no arquivos estáticos.
* **Leitura de logs:** Ao invés de usar uma consulta SQL, um arquivo de log ou mesmo a entrada padrão podem ser usados para fornecer os dados. Neste caso, qualquer filtragem deve ser aplicada já na entrada desses dados.

## Instalação e Uso

1. **Clone o repositório:**

    ```bash
    git clone https://github.com/hitalos/grapher.git
    cd grapher
    ```

2. **Crie um arquivo `.env`** na raiz do projeto com as seguintes variáveis de ambiente:

    ```env
    # String de conexão com o PostgreSQL (opcional)
    DSN="postgres://user:password@host:port/database"

    # Consulta SQL para buscar os dados (opcional)
    QUERY="SELECT time, value FROM sua_tabela WHERE time BETWEEN $1 AND $2 ORDER BY time"

    # Porta do servidor HTTP (opcional, padrão :6060).
    PORT=":6060"

    # Ambiente de desenvolvimento (opcional)
    ENV=dev
    ```

3. Caso a variável `QUERY` não seja fornecida, os dados serão lidos de um arquivo (primeiro argumento recebido pelo comando) ou da entrada padrão (caso nenhum arquivo seja informado). Será necessário definir as duas variáveis abaixo:

    ```env
    # Layout que será usado para interpretar o campo de data (no padrão da linguagem go)
    DT_LOG_FORMAT="Jan 02 15:04:05 2006"

    # Expressão regular usada para identificar os em cada linha (O trecho "?P<time>" será
    # usado para escolher o campo que será o "timestamp")
    LOG_REGEX="(?U)^\[[a-zA-Z]{3} (?P<time>.+)\] .+"
    ```

4. **Instale as dependências:**

    ```bash
    npm ci
    go mod tidy
    ```

5. **Execute a aplicação:**
    * Com `go run`:

        ```bash
        go run main.go
        ```

    * Com `make`:

        ```bash
        make
        ./dist/grapher
        ```

6. **Acesse a aplicação** em seu navegador: `http://localhost:6060`

## API

### `GET /data`

Retorna os dados da série temporal em formato JSON.

**Parâmetros:**

* `start` (opcional): Data de início no formato `YYYY-MM-DD`.
* `end` (opcional): Data de fim no formato `YYYY-MM-DD`.

**Exemplo:**

```bash
curl "http://localhost:8080/data?start=2024-01-01&end=2024-01-31"
```
