// Package dbstorage used for persisting urls in the database
package dbstorage

import (
	"context"
	"fmt"
	"runtime"
	"strings"
)

// DeleteURLBatch delete urls by ids
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
	sqlText := fmt.Sprintf(
		"UPDATE urls SET deleted_at = CURRENT_TIMESTAMP WHERE id in (%s)",
		strings.Trim(strings.Replace(fmt.Sprint(resIDs), " ", ",", -1), "[]"))

	_, err = s.db.ExecContext(ctx, sqlText)

	return err
}
