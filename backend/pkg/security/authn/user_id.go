package authn

import "fmt"

type UserID struct {
	IDInDomain string
	DomainID   string
}

func (userId *UserID) Equal(idInDomain, domainID string) bool {
	uid := *userId
	return idInDomain == uid.IDInDomain && domainID == uid.DomainID
}

func (userId *UserID) String() string {
	return fmt.Sprintf("%s@%s", userId.IDInDomain, userId.DomainID)
}
