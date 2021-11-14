CREATE TABLE IF NOT EXISTS accounts (
	id serial PRIMARY KEY,
	created_at date DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_at date DEFAULT CURRENT_TIMESTAMP NOT NULL,
	name text NOT NULL,
	cpf text UNIQUE NOT NULL,
	secret text NOT NULL,
	balance numeric DEFAULT 0 NOT NULL,
	active boolean DEFAULT false
);

CREATE TABLE IF NOT EXISTS transfers (
	id serial PRIMARY KEY,
	account_origin_id bigint REFERENCES accounts(id), 
	account_destination_id bigint REFERENCES accounts(id),
	amount numeric,
	created_at date DEFAULT CURRENT_TIMESTAMP NOT NULL
);