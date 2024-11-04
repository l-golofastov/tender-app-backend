package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"tender-app-backend/src/internal"
	"tender-app-backend/src/internal/config"
	"tender-app-backend/src/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func execCreateQuery(db *sql.DB, query string) error {
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	return nil
}

func New(cfg *config.Config) (*Storage, error) {
	const op = "storage.postgres.New"

	//connStr := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
	//	cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	connStr := cfg.ConnURL

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	createStatus := `
	CREATE TABLE IF NOT EXISTS status (
		id INT PRIMARY KEY,
		status_type VARCHAR(20)
	);`
	err = execCreateQuery(db, createStatus)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	addCreated, err := db.Prepare("INSERT INTO status(id, status_type) VALUES (1, 'CREATED') ON CONFLICT (id) DO NOTHING")
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	addPublished, err := db.Prepare("INSERT INTO status(id, status_type) VALUES (2, 'PUBLISHED') ON CONFLICT (id) DO NOTHING")
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	addClosed, err := db.Prepare("INSERT INTO status(id, status_type) VALUES (3, 'CLOSED') ON CONFLICT (id) DO NOTHING")
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	addCanceled, err := db.Prepare("INSERT INTO status(id, status_type) VALUES (4, 'CANCELED') ON CONFLICT (id) DO NOTHING")
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	_, err = addCreated.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	_, err = addPublished.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	_, err = addClosed.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	_, err = addCanceled.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	createTender := `
	CREATE TABLE IF NOT EXISTS tender(
	    id serial PRIMARY KEY,
	    name VARCHAR(100) NOT NULL,
	    description TEXT,
	    service_type VARCHAR(50),
	    status_id INT REFERENCES status(id) ON DELETE CASCADE,
	    organization_id INT REFERENCES organization(id) ON DELETE CASCADE,
	    creator_username VARCHAR(50) NOT NULL,
	    version INT NOT NULL
	);`
	err = execCreateQuery(db, createTender)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	createBid := `
	CREATE TABLE IF NOT EXISTS bid(
	    id serial PRIMARY KEY,
	    name VARCHAR(100) NOT NULL,
	    description TEXT,
	    status_id INT REFERENCES status(id) ON DELETE CASCADE,
	    tender_id INT REFERENCES organization_responsible_tender(id) ON DELETE CASCADE,
	    organization_id INT REFERENCES organization(id) ON DELETE CASCADE,
	    creator_username VARCHAR(50) NOT NULL,
	    version INT NOT NULL
	);`
	err = execCreateQuery(db, createBid)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	createOrgRespTender := `
	CREATE TABLE IF NOT EXISTS organization_responsible_tender(
	    id SERIAL PRIMARY KEY,
	    org_resp_id INT REFERENCES organization_responsible(id) ON DELETE CASCADE,
	    tender_id INT REFERENCES tender(id) ON DELETE CASCADE
	);`
	err = execCreateQuery(db, createOrgRespTender)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	createTenderBid := `
	CREATE TABLE IF NOT EXISTS tender_bid(
	    id SERIAL PRIMARY KEY,
	    tender_id INT REFERENCES organization_responsible_tender(id) ON DELETE CASCADE,
	    bid_id INT REFERENCES bid(id) ON DELETE CASCADE
	);`
	err = execCreateQuery(db, createTenderBid)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	createTenderVersions := `
	CREATE TABLE IF NOT EXISTS tender_versions(
	    id SERIAL PRIMARY KEY,
	    org_resp_tender_id INT REFERENCES organization_responsible_tender(id) ON DELETE CASCADE,
	    tender_id INT REFERENCES tender(id) ON DELETE CASCADE,
	    tender_version INT
	)`
	err = execCreateQuery(db, createTenderVersions)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	createBidVersions := `
	CREATE TABLE IF NOT EXISTS bid_versions(
	    id SERIAL PRIMARY KEY,
	    tender_bid_id INT REFERENCES tender_bid(id) ON DELETE CASCADE,
	    bid_id INT REFERENCES bid(id) ON DELETE CASCADE,
	    bid_version INT
	)`
	err = execCreateQuery(db, createBidVersions)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) GetOrgRespId(orgId int, creatorUsername string) (int, error) {
	const op = "storage.postgres.GetOrgRespId"

	stmt, err := s.db.Prepare(`
		SELECT r.id
		FROM organization_responsible AS r
		JOIN employee AS e
		ON r.user_id = e.id
		WHERE e.username = $1 AND r.organization_id = $2
	`)

	if err != nil {
		return 0, fmt.Errorf("%s %w", op, err)
	}

	var idResp int
	err = stmt.QueryRow(creatorUsername, orgId).Scan(&idResp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, storage.ErrOrgRespNotFound
		}
		return 0, fmt.Errorf("%s %w", op, err)
	}

	return idResp, nil
}

func (s *Storage) CreateTender(t internal.Tender) (internal.Tender, error) {
	const op = "storage.postgres.CreateTender"

	orgRespId, err := s.GetOrgRespId(t.OrganizationId, t.CreatorUsername)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	createTender, err := s.db.Prepare(`
		INSERT INTO tender(name, description, service_type, status_id, organization_id, creator_username, version)
		VALUES ($1, $2, $3, 1, $4, $5, 1) RETURNING id
	`)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	var tenderId int
	err = createTender.QueryRow(t.Name, t.Description, t.ServiceType, t.OrganizationId, t.CreatorUsername).Scan(&tenderId)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	orgRespTenderEntry, err := s.db.Prepare(`
		INSERT INTO organization_responsible_tender(org_resp_id, tender_id)
		VALUES ($1, $2) RETURNING id
	`)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	var id int
	err = orgRespTenderEntry.QueryRow(orgRespId, tenderId).Scan(&id)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	t.Id = id
	t.Version = 1
	t.Status = "CREATED"

	tenderVersionEntry, err := s.db.Prepare(`
		INSERT INTO tender_versions(org_resp_tender_id, tender_id, tender_version)
		VALUES ($1, $2, $3)
	`)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	_, err = tenderVersionEntry.Exec(id, tenderId, t.Version)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	return t, nil
}

func (s *Storage) PublishTender(id int) error {
	const op = "storage.postgres.PublishTender"

	stmt, err := s.db.Prepare(`
		UPDATE tender
		SET status_id = (SELECT id FROM status WHERE status_type='PUBLISHED')
		WHERE id=(SELECT tender_id FROM organization_responsible_tender WHERE id=$1)
	`)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	return nil
}

func (s *Storage) CloseTender(id int) error {
	const op = "storage.postgres.CloseTender"

	stmt, err := s.db.Prepare(`
		UPDATE tender
		SET status_id = (SELECT id FROM status WHERE status_type='CLOSED')
		WHERE id=(SELECT tender_id FROM organization_responsible_tender WHERE id=$1)
	`)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	return nil
}

func (s *Storage) GetTendersList() ([]internal.Tender, error) {
	const op = "storage.postgres.GetTendersList"

	stmt, err := s.db.Prepare(`
		SELECT r.id, t.name, t.description, t.service_type, s.status_type, t.organization_id, t.creator_username, t.version
		FROM organization_responsible_tender AS r JOIN tender AS t ON r.tender_id = t.id
		JOIN status as s ON t.status_id = s.id
	`)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	tenders := make([]internal.Tender, 0)

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	for rows.Next() {
		var t internal.Tender
		err = rows.Scan(&t.Id, &t.Name, &t.Description, &t.ServiceType, &t.Status, &t.OrganizationId, &t.CreatorUsername, &t.Version)
		if err != nil {
			return nil, fmt.Errorf("%s %w", op, err)
		}
		tenders = append(tenders, t)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	return tenders, nil
}

func (s *Storage) GetUserTendersList(username string) ([]internal.Tender, error) {
	const op = "storage.postgres.GetUserTendersList"

	stmt, err := s.db.Prepare(`
		SELECT r.id, t.name, t.description, t.service_type, s.status_type, t.organization_id, t.creator_username, t.version
		FROM organization_responsible_tender AS r JOIN tender AS t ON r.tender_id = t.id
		JOIN status as s ON t.status_id = s.id
		WHERE t.creator_username=$1
	`)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	tenders := make([]internal.Tender, 0)

	rows, err := stmt.Query(username)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	for rows.Next() {
		var t internal.Tender
		err = rows.Scan(&t.Id, &t.Name, &t.Description, &t.ServiceType, &t.Status, &t.OrganizationId, &t.CreatorUsername, &t.Version)
		if err != nil {
			return nil, fmt.Errorf("%s %w", op, err)
		}
		tenders = append(tenders, t)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	return tenders, nil
}

func (s *Storage) GetStatusId(status string) (int, error) {
	const op = "storage.postgres.GetStatusId"

	stmt, err := s.db.Prepare("SELECT id FROM status WHERE status_type = $1")
	if err != nil {
		return 0, fmt.Errorf("%s %w", op, err)
	}

	var id int

	err = stmt.QueryRow(status).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetTenderVersion(tenderId int) (int, error) {
	const op = "storage.postgres.GetTenderVersion"

	stmt, err := s.db.Prepare("SELECT MAX(tender_version) FROM tender_versions WHERE org_resp_tender_id = $1")
	if err != nil {
		return 0, fmt.Errorf("%s %w", op, err)
	}

	var version int

	err = stmt.QueryRow(tenderId).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("%s %w", op, err)
	}

	return version, nil
}

func (s *Storage) EditTender(t internal.Tender, editId int) (internal.Tender, error) {
	const op = "storage.postgres.EditTender"

	getTender, err := s.db.Prepare(`
		SELECT t.name, t.description, t.service_type, s.status_type, t.organization_id, t.creator_username, t.version
		FROM tender AS t JOIN organization_responsible_tender AS r ON t.id = r.tender_id
		JOIN status AS s ON t.status_id = s.id
		WHERE r.id = $1
	`)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	var edit internal.Tender

	err = getTender.QueryRow(editId).Scan(&edit.Name, &edit.Description, &edit.ServiceType, &edit.Status, &edit.OrganizationId, &edit.CreatorUsername, &edit.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return internal.Tender{}, storage.ErrTenderNotFound
		}

		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	actualVer, err := s.GetTenderVersion(editId)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	if t.Name != "" {
		edit.Name = t.Name
	}
	if t.Description != "" {
		edit.Description = t.Description
	}
	if t.ServiceType != "" {
		edit.ServiceType = t.ServiceType
	}

	edit.Id = editId
	edit.Version = actualVer + 1

	statusId, err := s.GetStatusId(edit.Status)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	createTender, err := s.db.Prepare(`
		INSERT INTO tender(name, description, service_type, status_id, organization_id, creator_username, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	var tenderId int

	err = createTender.QueryRow(edit.Name, edit.Description, edit.ServiceType, statusId, edit.OrganizationId, edit.CreatorUsername, edit.Version).Scan(&tenderId)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	updateOrgRespTend, err := s.db.Prepare(`
		UPDATE organization_responsible_tender
		SET tender_id = $1
		WHERE id = $2
	`)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	_, err = updateOrgRespTend.Exec(tenderId, editId)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	tenderVersionEntry, err := s.db.Prepare(`
		INSERT INTO tender_versions(org_resp_tender_id, tender_id, tender_version)
		VALUES ($1, $2, $3)
	`)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	_, err = tenderVersionEntry.Exec(editId, tenderId, edit.Version)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	return edit, nil
}

func (s *Storage) RollbackTender(tenderId, version int) (internal.Tender, error) {
	const op = "storage.postgres.RollbackTender"

	getPrev, err := s.db.Prepare(`
		SELECT t.name, t.description, t.service_type, s.status_type, t.organization_id, t.creator_username, t.version
		FROM tender_versions AS v JOIN tender AS t ON t.id = v.tender_id
		JOIN status AS s ON t.status_id = s.id
		WHERE v.org_resp_tender_id = $1 AND v.tender_version = $2
	`)
	if err != nil {
		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	var prev internal.Tender

	err = getPrev.QueryRow(tenderId, version).Scan(&prev.Name, &prev.Description, &prev.ServiceType, &prev.Status, &prev.OrganizationId, &prev.CreatorUsername, &prev.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return internal.Tender{}, storage.ErrTenderNotFound
		}

		return internal.Tender{}, fmt.Errorf("%s %w", op, err)
	}

	return s.EditTender(prev, tenderId)
}

func (s *Storage) CheckTenderExist(tenderId int) (bool, error) {
	const op = "storage.postgres.CheckTenderExist"

	stmt, err := s.db.Prepare("SELECT id FROM organization_responsible_tender WHERE id = $1")
	if err != nil {
		return false, fmt.Errorf("%s %w", op, err)
	}

	var res int

	err = stmt.QueryRow(tenderId).Scan(&res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrTenderNotFound
		}

		return false, fmt.Errorf("%s %w", op, err)
	}

	return true, nil
}

func (s *Storage) CreateBid(b internal.Bid) (internal.Bid, error) {
	const op = "storage.postgres.CreateBid"

	_, err := s.GetOrgRespId(b.OrganizationId, b.CreatorUsername)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	_, err = s.CheckTenderExist(b.TenderId)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	createBid, err := s.db.Prepare(`
		INSERT INTO bid(name, description, status_id, tender_id, organization_id, creator_username, version)
		VALUES ($1, $2, 1, $3, $4, $5, 1) RETURNING id
	`)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	var bidId int
	err = createBid.QueryRow(b.Name, b.Description, b.TenderId, b.OrganizationId, b.CreatorUsername).Scan(&bidId)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	tenderBidEntry, err := s.db.Prepare(`
		INSERT INTO tender_bid(tender_id, bid_id)
		VALUES ($1, $2) RETURNING id
	`)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	var id int
	err = tenderBidEntry.QueryRow(b.TenderId, bidId).Scan(&id)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	b.Id = id
	b.Version = 1
	b.Status = "CREATED"

	bidVersionEntry, err := s.db.Prepare(`
		INSERT INTO bid_versions(tender_bid_id, bid_id, bid_version)
		VALUES ($1, $2, $3)
	`)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	_, err = bidVersionEntry.Exec(id, bidId, b.Version)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	return b, nil
}

func (s *Storage) PublishBid(id int) error {
	const op = "storage.postgres.PublishBid"

	stmt, err := s.db.Prepare(`
		UPDATE bid
		SET status_id = (SELECT id FROM status WHERE status_type='PUBLISHED')
		WHERE id = (SELECT bid_id FROM tender_bid WHERE id=$1)
	`)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	return nil
}

func (s *Storage) CancelBid(id int) error {
	const op = "storage.postgres.CancelBid"

	stmt, err := s.db.Prepare(`
		UPDATE bid
		SET status_id = (SELECT id FROM status WHERE status_type='CANCELED')
		WHERE id = (SELECT bid_id FROM tender_bid WHERE id=$1)
	`)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	return nil
}

func (s *Storage) GetBidVersion(bidId int) (int, error) {
	const op = "storage.postgres.GetBidVersion"

	stmt, err := s.db.Prepare("SELECT MAX(bid_version) FROM bid_versions WHERE tender_bid_id = $1")
	if err != nil {
		return 0, fmt.Errorf("%s %w", op, err)
	}

	var version int

	err = stmt.QueryRow(bidId).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("%s %w", op, err)
	}

	return version, nil
}

func (s *Storage) EditBid(b internal.Bid, editId int) (internal.Bid, error) {
	const op = "storage.postgres.EditBid"

	getBid, err := s.db.Prepare(`
		SELECT b.name, b.description, s.status_type, b.tender_id, b.organization_id, b.creator_username, b.version
		FROM bid AS b JOIN tender_bid AS t ON b.id = t.bid_id
		JOIN status AS s ON b.status_id = s.id
		WHERE t.id = $1
	`)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	var edit internal.Bid

	err = getBid.QueryRow(editId).Scan(&edit.Name, &edit.Description, &edit.Status, &edit.TenderId, &edit.OrganizationId, &edit.CreatorUsername, &edit.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return internal.Bid{}, storage.ErrBidNotFound
		}

		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	actualVer, err := s.GetBidVersion(editId)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	if b.Name != "" {
		edit.Name = b.Name
	}
	if b.Description != "" {
		edit.Description = b.Description
	}

	edit.Id = editId
	edit.Version = actualVer + 1

	statusId, err := s.GetStatusId(edit.Status)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	createBid, err := s.db.Prepare(`
		INSERT INTO bid(name, description, status_id, tender_id, organization_id, creator_username, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	var bidId int

	err = createBid.QueryRow(edit.Name, edit.Description, statusId, edit.TenderId, edit.OrganizationId, edit.CreatorUsername, edit.Version).Scan(&bidId)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	updateTenderBid, err := s.db.Prepare(`
		UPDATE tender_bid
		SET bid_id = $1
		WHERE id = $2
	`)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	_, err = updateTenderBid.Exec(bidId, editId)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	bidVersionEntry, err := s.db.Prepare(`
		INSERT INTO bid_versions(tender_bid_id, bid_id, bid_version)
		VALUES ($1, $2, $3)
	`)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	_, err = bidVersionEntry.Exec(editId, bidId, edit.Version)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	return edit, nil
}

func (s *Storage) RollbackBid(bidId, version int) (internal.Bid, error) {
	const op = "storage.postgres.RollbackBid"

	getPrev, err := s.db.Prepare(`
		SELECT b.name, b.description, s.status_type, b.tender_id, b.organization_id, b.creator_username, b.version
		FROM bid_versions AS v JOIN bid AS b ON b.id = v.bid_id
		JOIN status AS s ON b.status_id = s.id
		WHERE v.tender_bid_id = $1 AND v.bid_version = $2
	`)
	if err != nil {
		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	var prev internal.Bid

	err = getPrev.QueryRow(bidId, version).Scan(&prev.Name, &prev.Description, &prev.Status, &prev.TenderId, &prev.OrganizationId, &prev.CreatorUsername, &prev.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return internal.Bid{}, storage.ErrBidNotFound
		}

		return internal.Bid{}, fmt.Errorf("%s %w", op, err)
	}

	return s.EditBid(prev, bidId)
}

func (s *Storage) GetUserBidsList(username string) ([]internal.Bid, error) {
	const op = "storage.postgres.GetUserBidsList"

	stmt, err := s.db.Prepare(`
		SELECT t.id, b.name, b.description, s.status_type, b.tender_id, b.organization_id, b.creator_username, b.version
		FROM tender_bid AS t JOIN bid AS b ON t.bid_id = b.id
		JOIN status as s ON b.status_id = s.id
		WHERE b.creator_username = $1
	`)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	bids := make([]internal.Bid, 0)

	rows, err := stmt.Query(username)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	for rows.Next() {
		var b internal.Bid
		err = rows.Scan(&b.Id, &b.Name, &b.Description, &b.Status, &b.TenderId, &b.OrganizationId, &b.CreatorUsername, &b.Version)
		if err != nil {
			return nil, fmt.Errorf("%s %w", op, err)
		}
		bids = append(bids, b)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	return bids, nil
}

func (s *Storage) GetTenderBidsList(tenderId int) ([]internal.Bid, error) {
	const op = "storage.postgres.GetTenderBidsList"

	_, err := s.CheckTenderExist(tenderId)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	stmt, err := s.db.Prepare(`
		SELECT t.id, b.name, b.description, s.status_type, b.tender_id, b.organization_id, b.creator_username, b.version
		FROM tender_bid AS t JOIN bid AS b ON t.bid_id = b.id
		JOIN status as s ON b.status_id = s.id
		WHERE b.tender_id = $1
	`)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	bids := make([]internal.Bid, 0)

	rows, err := stmt.Query(tenderId)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	for rows.Next() {
		var b internal.Bid
		err = rows.Scan(&b.Id, &b.Name, &b.Description, &b.Status, &b.TenderId, &b.OrganizationId, &b.CreatorUsername, &b.Version)
		if err != nil {
			return nil, fmt.Errorf("%s %w", op, err)
		}
		bids = append(bids, b)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	return bids, nil
}

func (s *Storage) CheckTenderPublished(id int) (bool, error) {
	const op = "storage.postgres.CheckTenderPublished"

	stmt, err := s.db.Prepare(`
		SELECT r.id
		FROM organization_responsible_tender AS r JOIN tender AS t ON r.tender_id = t.id
		JOIN status AS s ON t.status_id = s.id
		WHERE r.id = $1 AND s.status_type = 'PUBLISHED'
	`)
	if err != nil {
		return false, fmt.Errorf("%s %w", op, err)
	}

	var res int

	err = stmt.QueryRow(id).Scan(&res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrTenderNotPublished
		}

		return false, fmt.Errorf("%s %w", op, err)
	}

	return true, nil
}

func (s *Storage) CheckBidPublished(bidId int) (bool, error) {
	const op = "storage.postgres.CheckBidPublished"

	stmt, err := s.db.Prepare(`
		SELECT t.id
		FROM tender_bid AS t JOIN bid AS b ON t.bid_id = b.id
		JOIN status AS s ON b.status_id = s.id
		WHERE t.id = $1 AND s.status_type = 'PUBLISHED'
	`)
	if err != nil {
		return false, fmt.Errorf("%s %w", op, err)
	}

	var res int

	err = stmt.QueryRow(bidId).Scan(&res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrBidNotPublished
		}

		return false, fmt.Errorf("%s %w", op, err)
	}

	return true, nil
}

func (s *Storage) CheckBidExist(bidId int) (bool, error) {
	const op = "storage.postgres.CheckBidExist"

	stmt, err := s.db.Prepare("SELECT id FROM tender_bid WHERE id = $1")
	if err != nil {
		return false, fmt.Errorf("%s %w", op, err)
	}

	var res int

	err = stmt.QueryRow(bidId).Scan(&res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrBidNotFound
		}

		return false, fmt.Errorf("%s %w", op, err)
	}

	return true, nil
}

func (s *Storage) GetBidTenderId(bidId int) (int, error) {
	const op = "storage.postgres.GetBidTenderId"

	stmt, err := s.db.Prepare(`
		SELECT t.tender_id
		FROM tender_bid AS t JOIN bid AS b on t.bid_id = b.id
		WHERE t.id = $1
	`)
	if err != nil {
		return 0, fmt.Errorf("%s %w", op, err)
	}

	var tenderId int

	err = stmt.QueryRow(bidId).Scan(&tenderId)
	if err != nil {
		return 0, fmt.Errorf("%s %w", op, err)
	}

	return tenderId, nil
}

func (s *Storage) SubmitBid(bidId int, orgUsername string) error {
	const op = "storage.postgres.SubmitBid"

	_, err := s.CheckBidExist(bidId)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	_, err = s.CheckBidPublished(bidId)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	tenderId, err := s.GetBidTenderId(bidId)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	_, err = s.CheckTenderPublished(tenderId)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	checkOrg, err := s.db.Prepare(`
		SELECT r.id
		FROM organization_responsible AS r JOIN employee AS e ON r.user_id = e.id
		WHERE e.username = $1
		  AND r.organization_id = (SELECT t.organization_id
								   FROM organization_responsible_tender AS r JOIN tender AS t ON r.tender_id = t.id
								   WHERE r.id = $2)
	`)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	var orgId int

	err = checkOrg.QueryRow(orgUsername, tenderId).Scan(&orgId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrOrgRespNotFound
		}

		return fmt.Errorf("%s %w", op, err)
	}

	err = s.CancelBid(bidId)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	err = s.CloseTender(tenderId)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	return nil
}

func (s *Storage) GetBidsList() ([]internal.Bid, error) {
	const op = "storage.postgres.GetBidsList"

	stmt, err := s.db.Prepare(`
		SELECT t.id, b.name, b.description, s.status_type, b.tender_id, b.organization_id, b.creator_username, b.version
		FROM tender_bid AS t JOIN bid AS b ON t.bid_id = b.id
		JOIN status as s ON b.status_id = s.id
	`)
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	bids := make([]internal.Bid, 0)

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	for rows.Next() {
		var b internal.Bid
		err = rows.Scan(&b.Id, &b.Name, &b.Description, &b.Status, &b.TenderId, &b.OrganizationId, &b.CreatorUsername, &b.Version)
		if err != nil {
			return nil, fmt.Errorf("%s %w", op, err)
		}
		bids = append(bids, b)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}

	return bids, nil
}
