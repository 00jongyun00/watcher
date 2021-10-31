package dbrepo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tsawler/vigilate/internal/models"
)

// InsertHost inserts a host into the database
func (m *postgresDBRepo) InsertHost(h models.Host) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into hosts (host_name, canonical_name, url, ip, ipv6, location, os, activate, created_at, updated_at)
				values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id`

	var newID int

	err := m.DB.QueryRowContext(ctx, query,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		log.Println(err)
		return newID, err
	}

	// add host services and set inactive
	smtm := `
    insert into host_services (host_id, service_id, activate, schedule_number, schedule_unit,
    status, created_at, updated_at) values ($1, 1, 0, 3, 'm', 'pending', $2, $3)
  `
	_, err = m.DB.ExecContext(ctx, smtm, newID, time.Now(), time.Now())
	if err != nil {
		return newID, err
	}

	return newID, nil
}

func (m *postgresDBRepo) GetHostById(id int) (models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		select id, host_name, canonical_name, url, ip, ipv6, location, os, activate, created_at, updated_at
		from hosts where id = $1
`

	row := m.DB.QueryRowContext(ctx, query, id)

	var h models.Host

	err := row.Scan(
		&h.ID,
		&h.HostName,
		&h.CanonicalName,
		&h.URL,
		&h.IP,
		&h.IPV6,
		&h.Location,
		&h.OS,
		&h.Active,
		&h.CreatedAt,
		&h.UpdatedAt,
	)

	if err != nil {
		return h, err
	}

	return h, nil
}

func (m *postgresDBRepo) UpdateHost(h models.Host) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
    update hosts set host_name = $1, canonical_name = $2, url = $3, ip = $4, ipv6 = $5, os = $6,
    activate = $7, location = $8, updated_at = $9 where id = $10`

	_, err := m.DB.ExecContext(ctx, stmt,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.OS,
		h.Active,
		h.Location,
		time.Now(),
		h.ID,
	)

	if err != nil {
		fmt.Println("here")
		log.Println(err)
		return err
	}

	return nil
}

func (m *postgresDBRepo) AllHosts() ([]models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
    select id, host_name, canonical_name, url, ip, ipv6, location, os,
    activate, created_at, updated_at from hosts order by host_name
  `
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []models.Host

	for rows.Next() {
		var h models.Host
		err = rows.Scan(
			&h.ID,
			&h.HostName,
			&h.CanonicalName,
			&h.URL,
			&h.IP,
			&h.IPV6,
			&h.Location,
			&h.OS,
			&h.Active,
			&h.CreatedAt,
			&h.UpdatedAt,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		hosts = append(hosts, h)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return hosts, nil
}
