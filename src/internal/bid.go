package internal

type Bid struct {
	Id              int    `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	Status          string `json:"status,omitempty"`
	TenderId        int    `json:"tenderId,omitempty" validate:"required"`
	OrganizationId  int    `json:"organizationId,omitempty" validate:"required"`
	CreatorUsername string `json:"creatorUsername,omitempty" validate:"required"`
	Version         int    `json:"version,omitempty"`
}
