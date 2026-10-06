package database

import (
	"context"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pardnchiu/PodRun/internal/model"
)

func (s *SQLite) InsertRecord(ctx context.Context, d *model.Record) error {
	_, err := s.db.ExecContext(ctx, `
  INSERT INTO records (
    pod_id, content, hostname, ip
  )
  VALUES (
    (SELECT id FROM pods WHERE uid = ?), ?, ?, ?
  )
  `,
		d.UID, d.Content, d.Hostname, d.IP,
	)
	return err
}
