//go:generate mockgen -source=../../domain/interfaces.go -package=mocks -destination=mock_domain.go
//go:generate mockgen -source=../../datastore/db/mysql/repositories/interfaces.go -package=mocks -destination=mock_repos.go

package mocks
