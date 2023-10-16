package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/denisdubovitskiy/feedparser/internal/database"
	"github.com/denisdubovitskiy/feedparser/internal/unix"
)

func NewRunner(service *database.Service, maxRetries int64) *Runner {
	return &Runner{service: service, maxRetries: maxRetries}
}

type Runner struct {
	service    *database.Service
	maxRetries int64
}

func (r *Runner) ForEachSource(ctx context.Context, f func(source *database.Source) error) error {
	// Перед запуском сбрасываем количество ретраев.
	if err := r.service.ResetRetries(ctx); err != nil {
		return fmt.Errorf("runner: unable to reset retries: %v", err)
	}

	lastTimestamp, err := r.service.LastTimestamp(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Дата предыдущего прогона отсутствует.
			// Приложение инициализируется в первый раз.
			lastTimestamp = unix.TimeNow()
			if err := r.service.SetLastTimestamp(ctx, lastTimestamp); err != nil {
				return fmt.Errorf("runner: unable to initialize last timestamp: %v", err)
			}
		} else {
			// Другая непредвиденная ошибка
			return fmt.Errorf("runner: an error encountered while fetching a timestamp from the db: %v", err)
		}
	}

	// Время старта цикла опроса источников для последующей оценки длительности.
	jobStarted := time.Now()
	log.Printf("runner: starting at %s", jobStarted.Format(time.DateTime))

	unixStarted := lastTimestamp

	var finalErr error

	// Место, требующее ручного вмешательства.
	defer func() {
		unixFinished := unix.TimeNow()

		if err := r.service.SetLastTimestamp(context.Background(), unix.TimeNow()); err != nil {
			finalErr = fmt.Errorf("runner: unable to set timestamp %d: %v", unixFinished, err)
		}
	}()

	for {
		// Забираем из базы по одному источнику из тех, чья дата последнего
		// визита меньше, чем дата предыдущего запуска раннера.
		source, err := r.service.FetchOne(context.Background(), unixStarted, r.maxRetries)
		if err != nil {
			// Все источники пройдены.
			if errors.Is(err, sql.ErrNoRows) {
				break
			}

			log.Printf("runner: unable to fetch a source: %v", err)
			continue
		}

		if err := f(source); err != nil {
			log.Printf("runner: %s failed to process: %v", source.String(), err)

			retriesUpdateErr := r.service.UpdateRetries(context.Background(), source.ID)
			if retriesUpdateErr != nil {
				log.Printf("runner: %s failed to update retries: %v", source.String(), retriesUpdateErr)
			}

			continue
		}

		updateErr := r.service.UpdateLastVisited(context.Background(), source.ID, unix.TimeNow())
		if updateErr != nil {
			log.Printf("runner: %s update error: %v", source.String(), updateErr.Error())
			continue
		}

		log.Printf("runner: %s job finished", source.String())
	}

	log.Printf("runner: finished, time taken: %f", time.Since(jobStarted).Seconds())

	return finalErr
}
