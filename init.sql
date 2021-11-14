CREATE TABLE IF NOT EXISTS accounts (
	id serial PRIMARY KEY,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
	name text NOT NULL,
	cpf text UNIQUE NOT NULL,
	secret text NOT NULL,
	balance bigint DEFAULT 0 NOT NULL,
	active boolean DEFAULT true NOT NULL
);

CREATE TABLE IF NOT EXISTS transfers (
	id serial PRIMARY KEY,
	account_origin_id bigint REFERENCES accounts(id), 
	account_destination_id bigint REFERENCES accounts(id),
	amount bigint,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);