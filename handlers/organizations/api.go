package organizations

import (
	"net/http"
)

type OrganizationRoutes interface {
	CreateOrganization(w http.ResponseWriter, r *http.Request)
	GetOrganizations(w http.ResponseWriter, r *http.Request)
	GetOrganization(w http.ResponseWriter, r *http.Request)
	GetOrganizationMembers(w http.ResponseWriter, r *http.Request)
}
