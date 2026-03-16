package database

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func RegisterDBMetrics(db *sql.DB) {
	prometheus.MustRegister(collectors.NewDBStatsCollector(db, "recapo"))
}
