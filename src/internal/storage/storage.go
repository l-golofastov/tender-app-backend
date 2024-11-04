package storage

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrOrgRespNotFound    = errors.New("organisation responsible employee not found")
	ErrTenderNotFound     = errors.New("tndcreate with provided info not found")
	ErrBidNotFound        = errors.New("bid with provided info not found")
	ErrTenderNotPublished = errors.New("tndcreate not published")
	ErrBidNotPublished    = errors.New("bid not published")
)
