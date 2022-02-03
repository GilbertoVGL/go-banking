# Desafio Técnico - Go(lang)

Aplicação feita em Golang para o desafio técnico. 

### Considerações

Todas as rotas, exceto `GET /login` e `POST accounts` precisam de autenticação, sendo que a última deixei aberta para permitir fazer o fluxo completo da aplicação ao testá-la sem precisar fazer inserts no banco.

Todos os valores monetários são salvos em centavos (inteiros).

Deixei um .env já preenchido com os valores só para facilitar a execução do teste.


#### Rotas implementadas

##### `/`

- `GET /` - health check da aplicação

* * *

##### `/accounts`

- `GET /accounts` - obtém a lista de contas
- `GET /accounts/{account_id}/balance` - obtém o saldo da conta
- `GET /accounts/balance` - obtém o saldo da conta do usuário logado no momento
- `POST /accounts` - cria uma conta
  - body: `{
	    "name": "Roberval Neto",
      "cpf": "050.930.920-88",
	    "secret": "senha_segura",
	    "balance": 40
    }`

* * *

##### `/login`

- `POST /login` - autentica a usuaria
  - body: `{
	    "cpf": "610.781.580-53",
	    "secret": "lonedruid"
    }`

* * * 

##### `/transfers`

- `GET /transfers` - obtém a lista de transferencias da usuaria autenticada.
- `POST /transfers` - faz transferencia de uma conta para outra.
  - body:`{
	    "destination": 4,
      "amount": 1
    }`

* * *

## Rodando o APP

Para rodar o APP, caso ainda não exista, crie um arquivo .env seguindo o exemplo e preencha com os seus respectivos valores.

Para rodar usando docker: 
  - Executar o comando `docker-compose up`.

O comando vai iniciar a aplicação, subir o banco na porta `5432` e rodar o init.sql para criar as tabelas, e vai rodar um pgAdmin na porta `80`.
Ps.: É necessário rodar ao menos o banco de dados postgres no compose para evitar o trabalho de criar as tabelas manualmente.

Para rodar sem usar docker é preciso:
  - Ter uma instância do postgres9.6 rodando;
  - Criar uma database com o mesmo nome colocado na env `DB_NAME`;
  - Rodar o script init.sql. 
  - Executar o comando `go run cmd/server/main.go`.
 
## Rodando os testes

Na raíz do projeto execute: `go test ./...`. Atualmente somente os handlers http e o server possuem testes unitários.
