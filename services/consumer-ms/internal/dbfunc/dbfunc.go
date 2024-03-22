package dbfunc

import (
	"context"

	s "consumer-ms/internal/structs"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AddUser(pool *pgxpool.Pool, u *s.User) error {
	query := "INSERT INTO users(name, balance) VALUES($1,$2) RETURNING id;"
	row := pool.QueryRow(context.Background(), query, u.Name, u.Balance)
	err := row.Scan(&u.Id)
	if err != nil {
		return err
	}
	return nil
}

func Transaction(pool *pgxpool.Pool, from string, to string, val int) error {
	t, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	query := "UPDATE users SET balance = balance - $1 WHERE name=$2;"
	_, err = pool.Exec(context.Background(), query, val, from)
	if err != nil {
		return err
	}
	query = "UPDATE users SET balance = balance + $1 WHERE name=$2;"
	_, err = pool.Exec(context.Background(), query, val, to)
	if err != nil {
		return err
	}
	t.Commit(context.Background())
	return nil
}

func DeleteUser(pool *pgxpool.Pool, name string) error {
	query := "DELETE FROM users WHERE name=$1;"
	_, err := pool.Exec(context.Background(), query, name)
	if err != nil {
		return err
	}
	return nil
}

func GetUserByName(pool *pgxpool.Pool, name string) (*s.User, error) {
	query := "SELECT * FROM users WHERE name=$1;"
	row := pool.QueryRow(context.Background(), query, name)
	result := s.User{}
	err := row.Scan(&result.Id, &result.Name, &result.Balance)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
