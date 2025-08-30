package worker

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func startMaterializedViewRefresher(db *gorm.DB, logger *logrus.Logger, viewName string) *gocron.Scheduler {
	// TODO: Add distributed lock
	s := gocron.NewScheduler(time.Local)

	if _, err := s.Every(10).Minutes().Do(func() {
		logger.Infof("Refreshing materialized view: %s", viewName)
		if err := db.Exec(fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", viewName)).Error; err != nil {
			logger.Errorf("Error refreshing materialized view %s: %v", viewName, err)
		} else {
			logger.Infof("Successfully refreshed materialized view: %s", viewName)
		}
	}); err != nil {
		logger.Fatalf("Failed to schedule materialized view refresher: %v", err)
	}

	s.StartAsync()
	return s
}
