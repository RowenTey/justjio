package workers

import (
	"fmt"
	"os"

	"github.com/RowenTey/JustJio/server/api/pkg/app"
	gormlock "github.com/go-co-op/gocron-gorm-lock/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/sirupsen/logrus"
)

func startMaterializedViewRefresher(ctx *app.Context, viewName string) *gocron.Scheduler {
	logger := ctx.Logger.WithFields(logrus.Fields{"component": "MaterializedViewRefresher"})
	db := ctx.DB

	hostname, err := os.Hostname()
	if err != nil {
		logger.Fatalf("Failed to get hostname: %v", err)
	}
	logger.Infof("Hostname: %s", hostname)

	locker, err := gormlock.NewGormLocker(db, hostname)
	if err != nil {
		logger.Fatalf("Failed to create DB locker: %v", err)
	}

	s, err := gocron.NewScheduler(gocron.WithDistributedLocker(locker))
	if err != nil {
		logger.Fatalf("Failed to create scheduler: %v", err)
	}

	refreshSql := fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", viewName)
	refreshTask := func() {
		logger.Debugf("Refreshing materialized view: %s", viewName)
		if err := db.Exec(refreshSql).Error; err != nil {
			logger.Errorf("Error refreshing materialized view %s: %v", viewName, err)
		}
		logger.Debugf("Successfully refreshed materialized view: %s", viewName)
	}

	logger.Info("Starting materialized view refresher...")
	if _, err := s.NewJob(
		gocron.CronJob("*/1 * * * *", false),
		gocron.NewTask(refreshTask),
	); err != nil {
		logger.Fatalf("Failed to schedule materialized view refresher: %v", err)
	}

	s.Start()
	return &s
}
