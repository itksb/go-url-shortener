package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// SaveURL - saves url to the postgres db
func (s *Storage) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	var err error
	err = s.reconnect(ctx)
	if err != nil {
		s.l.Error(err)
		return "", err
	}

	query := `INSERT INTO urls (user_id, original_url) VALUES ($1, $2)
              ON CONFLICT ON CONSTRAINT urls_unique_idx DO NOTHING RETURNING id`
	row := s.db.QueryRowContext(ctx, query, userID, url)

	var ID int
	err = row.Scan(&ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			//query does not return id, so duplicate conflict, need to retrieve id from db
			row := s.db.QueryRowContext(ctx, `SELECT id FROM urls WHERE original_url = $1`, url)
			err = row.Scan(&ID)
			if err != nil {
				s.l.Error(err)
				return "", err
			}
			return fmt.Sprint(ID), fmt.Errorf("%w", shortener.ErrDuplicate)
		}

	}

	return fmt.Sprint(ID), nil
}

// GetURL - retrieves url from the underlying db by id
func (s *Storage) GetURL(ctx context.Context, id string) (shortener.URLListItem, error) {
	result := shortener.URLListItem{}
	var err error
	err = s.reconnect(ctx)
	if err != nil {
		s.l.Error(err)
		return result, err
	}

	query := `SELECT id, user_id, original_url, deleted_at FROM urls WHERE id = $1`

	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return result, err
	}
	res := s.db.QueryRowContext(ctx, query, idInt64)
	err = res.Scan(&result.ID, &result.UserID, &result.OriginalURL, &result.DeletedAt)
	if err != nil {
		s.l.Error(err)
		return result, err
	}

	return result, nil
}

// ListURLByUserID - list urls by user
func (s *Storage) ListURLByUserID(ctx context.Context, userID string) ([]shortener.URLListItem, error) {
	urls := []shortener.URLListItem{}
	var err error
	err = s.reconnect(ctx)
	if err != nil {
		s.l.Error(err)
		return urls, err
	}

	query := "SELECT * FROM urls WHERE user_id=$1"

	err = s.db.Select(&urls, query, userID)
	if err != nil {
		s.l.Error(err)
		return urls, err
	}

	return urls, nil

}

func (s *Storage) DeleteURLBatch(ctx context.Context, userID string, ids []string) error {
	var err error
	err = s.reconnect(ctx)
	if err != nil {
		s.l.Error(err)
		return err
	}

	inputCh := make(chan string)

	// send input ids to the inputCh
	go func() {
		for i := 0; i < len(ids); i++ {
			inputCh <- ids[i]
		}
		close(inputCh)
	}()

	// здесь fanOut
	workersCount := runtime.NumCPU()
	fanOutChs := fanOut(inputCh, workersCount)

	workerChs := make([]chan int64, 0, workersCount)
	for _, fanOutCh := range fanOutChs {
		workerCh := make(chan int64)
		newWorker(fanOutCh, workerCh)
		workerChs = append(workerChs, workerCh)
	}

	resIDs := make([]int64, len(ids))
	// здесь fanIn
	for v := range fanIn(workerChs...) {
		resIDs = append(resIDs, v)
	}
	sql := fmt.Sprintf(
		"UPDATE urls SET deleted_at = CURRENT_TIMESTAMP WHERE id in (%s)",
		strings.Trim(strings.Replace(fmt.Sprint(resIDs), " ", ",", -1), "[]"))

	_, err = s.db.ExecContext(ctx, sql)

	return err
}

func newWorker(input chan string, out chan int64) {
	go func() {
		for val := range input {
			id, err := strconv.ParseInt(val, 10, 64)
			if err == nil {
				out <- id
			}
		}
		close(out)
	}()
}

// Разделим изначальный канал на N каналов, где N равно числу воркеров, которые будут обрабатывать данные.
// Для этого создадим слайс из N каналов, куда будем раскладывать данные в отдельной горутине по принципу round-robin.
// Когда родительский канал будет закрыт, горутина завершит работу.
func fanOut(inputCh chan string, n int) []chan string {
	chs := make([]chan string, 0, n)
	for i := 0; i < n; i++ {
		ch := make(chan string)
		chs = append(chs, ch)
	}

	go func() {
		defer func(chs []chan string) {
			for _, ch := range chs {
				close(ch)
			}
		}(chs)

		for i := 0; ; i++ {
			if i == len(chs) {
				i = 0
			}

			val, ok := <-inputCh
			if !ok {
				return
			}

			ch := chs[i]
			ch <- val
		}
	}()

	return chs
}

// Напишем fanIn-функцию. Она будет принимать входные каналы как variadic-аргумент,
// а затем запускать по одной горутине для каждого входного канала.
// Горутина будет перенаправлять вычитанные из входного канала данные в выходной канал.
// Чтобы вести учёт запущенных горутин, используем sync.WaitGroup и заблокируемся на wg.Wait.
// Тогда выходной канал закроется только после того, как закроются все входные каналы.
func fanIn(inputChs ...chan int64) chan int64 {
	outCh := make(chan int64)
	go func() {
		wg := &sync.WaitGroup{}

		for _, inputCh := range inputChs {
			wg.Add(1)

			go func(inputCh chan int64) {
				defer wg.Done()
				for item := range inputCh {
					outCh <- item
				}
			}(inputCh)
		}
		wg.Wait()
		close(outCh)
	}()

	return outCh
}
