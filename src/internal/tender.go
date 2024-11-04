package internal

type Tender struct {
	Id              int    `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	ServiceType     string `json:"serviceType,omitempty"`
	Status          string `json:"status,omitempty"`
	OrganizationId  int    `json:"organizationId,omitempty" validate:"required"`
	CreatorUsername string `json:"creatorUsername,omitempty" validate:"required"`
	Version         int    `json:"version,omitempty"`
}
