package organizations

import (
	"net/http"
)

type OrganizationRoutes interface {
	GetOrganizations(w http.ResponseWriter, r *http.Request)
	GetOrganization(w http.ResponseWriter, r *http.Request)
	CreateOrganization(w http.ResponseWriter, r *http.Request)
}
