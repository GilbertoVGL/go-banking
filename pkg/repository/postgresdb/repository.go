package postgresdb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

	pgx "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/apperrors"
	"github.com/GilbertoVGL/go-banking/pkg/login"
	"github.com/GilbertoVGL/go-banking/pkg/transfer"
)

type postgresDB struct {
	db *pgxpool.Pool
}

func new() (*pgxpool.Pool, error) {
	dsn := url.URL{
		User:     url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PW")),
		Scheme:   "postgres",
		Host:     fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		Path:     os.Getenv("DB_NAME"),
		RawQuery: (&url.Values{"sslmode": []string{"disable"}}).Encode(),
	}

	config, err := pgxpool.ParseConfig(dsn.String())

	if err != nil {
		return nil, err
	}

	pool, err := strconv.Atoi(os.Getenv("DB_MAX_POOL"))

	if err != nil {
		return nil, err
	}

	config.MaxConns = int32(pool)

	db, err := pgxpool.ConnectConfig(context.Background(), config)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *postgresDB) getConn() (*pgxpool.Conn, error) {
	if r == nil || r.db == nil {
		db, err := new()
		if err != nil {
			log.Print("warning: unable to initialize db:", err)
			return nil, apperrors.NewDatabaseError("unable to initialize db", err.Error())
		}
		r = &postgresDB{db}
	}

	conn, err := r.db.Acquire(context.Background())

	if err != nil {
		return conn, apperrors.NewDatabaseError("unable to get database", err.Error())
	}

	return conn, nil
}

func New() (*postgresDB, error) {
	db, err := new()
	if err != nil {
		return &postgresDB{}, apperrors.NewDatabaseError("unable to initialize db", err.Error())
	}

	return &postgresDB{db}, nil
}

func (r *postgresDB) Close() {
	r.db.Close()
}

func (r *postgresDB) GetAccountById(ctx context.Context, id uint64) (account.Account, error) {
	var account account.Account

	select {
	default:
		conn, err := r.getConn()

		if err != nil {
			return account, err
		}

		defer conn.Release()

		query := fmt.Sprintf("select id, name, cpf, balance, active from accounts where id = '%d';", id)

		if err := conn.QueryRow(ctx, query).Scan(&account.Id, &account.Name, &account.Cpf, &account.Balance, &account.Active); err != nil {
			if errors.Is(pgx.ErrNoRows, err) {
				return account, apperrors.NewAccountNotFoundError("account not found")
			}
			return account, apperrors.NewDatabaseError(err.Error())
		}

		return account, nil
	case <-ctx.Done():
		return account, ctx.Err()
	}
}

func (r *postgresDB) GetAccountBySecretAndCPF(ctx context.Context, l login.LoginRequest) (login.Account, error) {
	var account login.Account
	select {
	default:
		conn, err := r.getConn()

		if err != nil {
			return account, err
		}

		defer conn.Release()
		query := fmt.Sprintf("select id, active from accounts where cpf = '%s' AND secret = '%s';", l.Cpf, l.Secret)

		if err := conn.QueryRow(ctx, query).Scan(&account.Id, &account.Active); err != nil {
			if errors.Is(pgx.ErrNoRows, err) {
				return account, apperrors.NewDatabaseError("invalid cpf or password")
			}

			return account, err
		}

		return account, nil
	case <-ctx.Done():
		return account, ctx.Err()
	}
}

func (r *postgresDB) ListAccount(ctx context.Context, params account.ListAccountQuery) (account.ListAccountsReponse, error) {
	var accountsResponse account.ListAccountsReponse

	select {
	default:
		conn, err := r.getConn()

		if err != nil {
			return accountsResponse, err
		}

		defer conn.Release()

		query := fmt.Sprintf(`select 
								id, 
								name, 
								cpf, 
								balance 
							from 
								accounts 
							order by id 
							limit %d 
							offset %d;`, params.PageSize, (params.PageSize * params.Page))
		rows, err := conn.Query(ctx, query)

		if err != nil && !errors.Is(pgx.ErrNoRows, err) {
			return accountsResponse, apperrors.NewDatabaseError(err.Error())
		}

		accounts := []account.ListAccount{}

		for rows.Next() {
			var account account.ListAccount

			if err := rows.Scan(&account.Id, &account.Name, &account.Cpf, &account.Balance); err != nil {
				return accountsResponse, apperrors.NewDatabaseError(err.Error())
			}

			accounts = append(accounts, account)
		}

		var count int64
		countQuery := fmt.Sprintf("select count(*) from accounts;")

		if err := conn.QueryRow(ctx, countQuery).Scan(&count); err != nil {
			return accountsResponse, apperrors.NewDatabaseError(err.Error())
		}

		accountsResponse.Data = accounts
		accountsResponse.Total = count
		accountsResponse.Page = int64(params.Page + 1)

		return accountsResponse, nil
	case <-ctx.Done():
		return accountsResponse, ctx.Err()
	}
}

func (r *postgresDB) AddAccount(ctx context.Context, a account.NewAccountRequest) error {
	select {
	default:
		conn, err := r.getConn()

		if err != nil {
			return err
		}

		defer conn.Release()

		query := fmt.Sprintf("insert into accounts (name, cpf, balance, secret) values ('%s', '%s', %d, '%s')", a.Name, a.Cpf, a.Balance, a.Secret)
		_, err = conn.Exec(ctx, query)

		if err != nil {
			return apperrors.NewDatabaseError(err.Error())
		}

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *postgresDB) GetAccountBalance(ctx context.Context, id uint64) (int64, error) {
	var balance int64

	select {
	default:
		conn, err := r.getConn()

		if err != nil {
			return balance, err
		}

		defer conn.Release()

		query := fmt.Sprintf("select balance from accounts where id = %d", id)
		if err := conn.QueryRow(ctx, query).Scan(&balance); err != nil {
			return balance, apperrors.NewDatabaseError(err.Error())
		}

		return balance, nil
	case <-ctx.Done():
		return balance, ctx.Err()
	}
}

func (r *postgresDB) GetTransfers(ctx context.Context, id uint64, params transfer.ListTransferQuery) (transfer.ListTransferResponse, error) {
	var transferResponse transfer.ListTransferResponse

	select {
	default:
		conn, err := r.getConn()

		if err != nil {
			return transferResponse, err
		}

		defer conn.Release()

		query := fmt.Sprintf(`select 
							tr.amount,
							tr.created_at,
							oa.name,
							oa.cpf,
							da.name,
							da.cpf
						from transfers as tr
						inner join accounts as oa
							on tr.account_origin_id = oa.id
						inner join accounts as da
							on tr.account_destination_id = da.id
						where 
							tr.account_origin_id = %d 
							or 
							tr.account_destination_id = %d
						order by tr.id 
						limit %d 
						offset %d;`, id, id, params.PageSize, (params.PageSize * params.Page))
		rows, err := conn.Query(ctx, query)

		if err != nil && !errors.Is(pgx.ErrNoRows, err) {
			return transferResponse, apperrors.NewDatabaseError(err.Error())
		}

		accounts := []transfer.ListTransfer{}

		for rows.Next() {
			var transfer transfer.ListTransfer

			if err := rows.Scan(&transfer.Amount, &transfer.CreatedAt, &transfer.OriginName, &transfer.OriginCpf, &transfer.DestinationName, &transfer.DestinationCpf); err != nil {

				return transferResponse, apperrors.NewDatabaseError(err.Error())
			}

			accounts = append(accounts, transfer)
		}

		if err != nil {
			return transferResponse, apperrors.NewDatabaseError(err.Error())
		}

		var count int64
		countQuery := fmt.Sprintf(`select 
									count(*)
								from transfers as tr
								inner join accounts as oa
									on tr.account_origin_id = oa.id
								inner join accounts as da
									on tr.account_destination_id = da.id
								where 
									tr.account_origin_id = %d 
									or 
									tr.account_destination_id = %d;`, id, id)

		if err := conn.QueryRow(ctx, countQuery).Scan(&count); err != nil {
			return transferResponse, apperrors.NewDatabaseError(err.Error())
		}

		transferResponse.Data = accounts
		transferResponse.Total = count
		transferResponse.Page = int64(params.Page + 1)

		return transferResponse, nil
	case <-ctx.Done():
		return transferResponse, ctx.Err()
	}
}

func (r *postgresDB) AddTransfer(ctx context.Context, t transfer.TransferRequest) error {
	select {
	default:
		conn, err := r.getConn()

		if err != nil {
			return err
		}

		defer conn.Release()

		tx, err := conn.Begin(ctx)

		if err != nil {
			return apperrors.NewDatabaseError(err.Error())
		}

		defer tx.Rollback(ctx)

		insertTransferQuery := fmt.Sprintf("insert into transfers (account_origin_id, account_destination_id, amount) values ('%d', '%d', %d)", t.Origin, *t.Destination, *t.Amount)
		originBalanceQuery := fmt.Sprintf("update accounts set balance = balance - %d where id = %d", *t.Amount, t.Origin)
		destinationBalanceQuery := fmt.Sprintf("update accounts set balance = balance + %d where id = %d", *t.Amount, *t.Destination)

		if _, err = tx.Exec(ctx, insertTransferQuery); err != nil {
			return apperrors.NewDatabaseError(err.Error())
		}

		if _, err = tx.Exec(ctx, originBalanceQuery); err != nil {
			return apperrors.NewDatabaseError(err.Error())
		}

		if _, err = tx.Exec(ctx, destinationBalanceQuery); err != nil {
			return apperrors.NewDatabaseError(err.Error())
		}

		err = tx.Commit(ctx)

		if err != nil {
			return apperrors.NewDatabaseError(err.Error())
		}

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
