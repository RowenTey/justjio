package workers

import (
	"fmt"
	"time"

	"github.com/RowenTey/JustJio/server/api/pkg/app"
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

func startMaterializedViewRefresher(ctx *app.Context, viewName string) *gocron.Scheduler {
	logger := ctx.Logger.WithFields(logrus.Fields{"component": "MaterializedViewRefresher"})
	db := ctx.DB

	// TODO: Add distributed lock
	s := gocron.NewScheduler(time.Local)

	refreshSql := fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", viewName)

	refreshTask := func() {
		logger.Debugf("Refreshing materialized view: %s", viewName)
		if err := db.Exec(refreshSql).Error; err != nil {
			logger.Errorf("Error refreshing materialized view %s: %v", viewName, err)
		}
		logger.Debugf("Successfully refreshed materialized view: %s", viewName)
	}

	logger.Info("Starting materialized view refresher...")
	if _, err := s.Every(1).Minutes().Do(refreshTask); err != nil {
		logger.Fatalf("Failed to schedule materialized view refresher: %v", err)
	}

	s.StartAsync()
	return s
}
