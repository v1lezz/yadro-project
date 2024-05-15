package index

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"yadro-project/internal/config"
	"yadro-project/internal/core/ports"
	"yadro-project/pkg/pair"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type PostgresConn struct {
	pool *pgxpool.Pool
}

func NewPostgresConn(ctx context.Context, cfg config.PostgresDBConfig) (*PostgresConn, error) {
	pgCFG, err := pgxpool.ParseConfig(cfg.String())
	if err != nil {
		return nil, fmt.Errorf("error parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgCFG)
	if err != nil {
		return nil, fmt.Errorf("error create new postgres pool")
	}

	return &PostgresConn{
		pool: pool,
	}, nil
}

const getKeywords = `SELECT keyword FROM keyword`
const getComics = `SELECT index.comics_id FROM keyword INNER JOIN index ON keyword.id = index.keyword_id WHERE keyword = $1`

func (pg *PostgresConn) GetNumbersOfNMostRelevantComics(ctx context.Context, n int, keywords []string) ([]int, error) {
	rows, err := pg.pool.Query(ctx, getKeywords)
	if err != nil {
		return nil, err
	}

	k := make(map[string][]int)

	var temp string
	for rows.Next() {
		if err := rows.Scan(&temp); err != nil {
			return nil, fmt.Errorf("error scan: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	rows.Close()

	for keyword, _ := range k {
		rows, err = pg.pool.Query(ctx, getComics, keyword)
		if err != nil {
			return nil, err
		}

		IDs := make([]int, 0)
		var ID int

		for rows.Next() {
			if err := rows.Scan(&ID); err != nil {
				return nil, fmt.Errorf("error scan: %w", err)
			}
			IDs = append(IDs, ID)
		}

		rows.Close()
		k[keyword] = IDs
	}

	cnt := make(map[int]int)
	for _, keyword := range keywords {
		for _, number := range k[keyword] {
			cnt[number]++
		}
	}

	return pair.GetNRelevantFromMap(cnt, n), nil
}

const insertKeyword = `INSERT INTO keyword(keyword) VALUES ($1) ON CONFLICT (keyword) DO NOTHING`
const insertKeywordComics = `INSERT INTO index(keyword_id, comics_id) VALUES ((SELECT id FROM keyword WHERE keyword = $1), $2)`

func (pg *PostgresConn) UpdateIndex(ctx context.Context, id int, keywords []string) error {
	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error create transaction")
	}
	defer tx.Rollback(ctx)

	g, gCtx := errgroup.WithContext(ctx)
	for _, keyword := range keywords {
		g.Go(func() error {
			if _, err := tx.Exec(gCtx, insertKeyword, keyword); err != nil {
				return err
			}

			if _, err := tx.Exec(gCtx, insertKeywordComics, keyword, id); err != nil {
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return err
}

const updateLastUpdateTime = `UPDATE time SET update_time_index = $1 WHERE id = 1`

func (pg *PostgresConn) Save(ctx context.Context, updateTime time.Time) error {
	if _, err := pg.pool.Exec(ctx, updateLastUpdateTime, updateTime); err != nil {
		return fmt.Errorf("error update last update time: %w", err)
	}
	pg.pool.Close()
	return nil
}

const getLastUpdateTime = `SELECT update_time_index FROM time WHERE id = 1`

func (pg *PostgresConn) GetLastUpdateTime(ctx context.Context) (time.Time, error) {
	row := pg.pool.QueryRow(ctx, getLastUpdateTime)

	t := time.Time{}
	if err := row.Scan(&t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, ports.ErrIsNotExist
		}
		return time.Time{}, fmt.Errorf("error get last update time: %w", err)
	}

	return t, nil
}

const clear = `TRUNCATE TABLE index`

func (pg *PostgresConn) Clear(ctx context.Context) error {
	if _, err := pg.pool.Exec(ctx, clear); err != nil {
		return fmt.Errorf("error clear index")
	}
	return nil
}
