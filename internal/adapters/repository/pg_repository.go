package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"yadro-project/internal/adapters/repository/migrations"
	"yadro-project/internal/config"
	"yadro-project/internal/core/domain"
	"yadro-project/internal/core/ports"

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

	if err = migrations.Up(pool, "/server/internal/adapters/repository/migrations"); err != nil {
		return nil, fmt.Errorf("error up migrations: %w", err)
	}

	return &PostgresConn{
		pool: pool,
	}, nil
}

const getComics = `SELECT * FROM comics`

func (pg *PostgresConn) GetComics(ctx context.Context) ([]domain.Comics, error) {
	rows, err := pg.pool.Query(ctx, getComics)
	if err != nil {
		return nil, fmt.Errorf("error get query of comics: %w", err)
	}

	comics := make([]domain.Comics, 0)
	for rows.Next() {
		c := domain.Comics{}
		if err = rows.Scan(&c.ID, &c.ImgURL, &c.Keywords); err != nil {
			return nil, fmt.Errorf("error scan row: %w", err)
		}
		comics = append(comics, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error rows: %w", err)
	}

	for idx, c := range comics {
		keywords, err := pg.getKeywordById(ctx, c.ID)
		if err != nil {
			return nil, fmt.Errorf("eror get keywords by id: %w", err)
		}
		comics[idx].Keywords = keywords
	}
	return comics, nil
}

const getKeywordById = `SELECT keyword FROM keyword INNER JOIN comics_keyword ON keyword.id = comics_keyword.keyword_id WHERE comics_keyword.comics_id = $1`

func (pg *PostgresConn) getKeywordById(ctx context.Context, id int) ([]string, error) {
	rows, err := pg.pool.Query(ctx, getKeywordById, id)
	if err != nil {
		return nil, fmt.Errorf("error query: %w", err)
	}

	defer rows.Close()

	ans := make([]string, 0)
	var temp string
	for rows.Next() {
		if err = rows.Scan(&temp); err != nil {
			return nil, fmt.Errorf("error scan: %w", err)
		}
		ans = append(ans, temp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error rows: %w", err)
	}
	return ans, nil
}

const getCountComics = `SELECT COUNT(*) FROM comics`

func (pg *PostgresConn) GetCountComics(ctx context.Context) (int, error) {
	row := pg.pool.QueryRow(ctx, getCountComics)
	cnt := 0
	if err := row.Scan(&cnt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("error get count comics: %w", err)
	}
	return cnt, nil
}

const getExistIDs = `SELECT id FROM comics`

func (pg *PostgresConn) GetIDMissingComics(ctx context.Context, cntInServer int) ([]int, error) {
	rows, err := pg.pool.Query(ctx, getExistIDs)
	if err != nil {
		return nil, fmt.Errorf("error get IDs: %w", err)
	}

	defer rows.Close()

	ids := make(map[int]bool)
	var id int
	for rows.Next() {
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("error scan: %w", err)
		}

		ids[id] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error rows: %w", err)
	}

	missingIDs := make([]int, 0, cntInServer-len(ids))

	for i := 1; i < cntInServer; i++ {
		if !ids[i] {
			missingIDs = append(missingIDs, i)
		}
	}

	return missingIDs, nil
}

const comicsIsExist = `SELECT 1 FROM comics WHERE id = $1`

func (pg *PostgresConn) comicsIsExist(ctx context.Context, id int) (bool, error) {
	row := pg.pool.QueryRow(ctx, comicsIsExist, id)

	if err := row.Scan(new(int)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("error check comics is exist: %w", err)
	}
	return true, nil
}

const (
	insertKeyword       = `INSERT INTO keyword(keyword) VALUES ($1) ON CONFLICT (keyword) DO NOTHING`
	insertComics        = `INSERT INTO comics VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`
	insertComicsKeyword = `INSERT INTO comics_keyword(comics_id, keyword_id) VALUES ($1, (SELECT keyword.id FROM keyword WHERE keyword.keyword = $2))`
)

func (pg *PostgresConn) Add(ctx context.Context, comics domain.Comics, id int) error {
	isExist, err := pg.comicsIsExist(ctx, comics.ID)
	if err != nil {
		return err
	}

	if !isExist {
		return ports.ErrIsNotExist
	}

	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error create tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, insertComics, comics.ID, comics.ImgURL); err != nil {
		return fmt.Errorf("error insert comics: %w", err)
	}

	g, gCtx := errgroup.WithContext(ctx)

	for _, keyword := range comics.Keywords {
		g.Go(func() error {
			if _, err := tx.Exec(gCtx, insertKeyword, keyword); err != nil {
				return err
			}
			if _, err := tx.Exec(gCtx, insertComicsKeyword, comics.ID, keyword); err != nil {
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error add keyword: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error commit: %w", err)
	}
	return nil
}

const updateLastUpdateTime = `UPDATE time SET update_time_comics = $1 WHERE id = 1`

func (pg *PostgresConn) Close(ctx context.Context, updateTime time.Time) error {
	if _, err := pg.pool.Exec(ctx, updateLastUpdateTime, updateTime); err != nil {
		return fmt.Errorf("error update last update time: %w", err)
	}
	pg.pool.Close()
	return nil
}

const getLastFulLCheckTime = `SELECT last_full_check_time FROM time WHERE id = 1`

func (pg *PostgresConn) GetLastFullCheckTime(ctx context.Context) (time.Time, error) {
	row := pg.pool.QueryRow(ctx, getLastFulLCheckTime)

	t := time.Time{}
	if err := row.Scan(&t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, ports.ErrIsNotExist
		}
		return time.Time{}, fmt.Errorf("error get last update time: %w", err)
	}

	return t, nil

}

const updateLastFullCheckTime = `UPDATE time SET last_full_check_time = $1 WHERE id = 1`

func (pg *PostgresConn) UpdateLastFullCheckTime(ctx context.Context, updateTime time.Time) error {
	if _, err := pg.pool.Exec(ctx, updateLastFullCheckTime, updateTime); err != nil {
		return fmt.Errorf("error update last full check time: %w", err)
	}
	return nil
}

const getLastUpdateTime = `SELECT update_time_comics FROM time WHERE id = 1`

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

const getURLComicsByID = `SELECT image_url FROM comics WHERE id = $1`

func (pg *PostgresConn) GetURLComicsByID(ctx context.Context, ID int) (string, error) {
	row := pg.pool.QueryRow(ctx, getURLComicsByID, ID)

	var url string
	if err := row.Scan(&url); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ports.ErrIsNotExist
		}
		return "", fmt.Errorf("error get url comics by id: %w", err)
	}

	return url, nil
}
