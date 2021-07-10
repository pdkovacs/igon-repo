package authn

import "fmt"

// TODO: having the domain ID in the user ID is a nice idea in support of multi-domain user database and
// the initial implementation is also very promising (creates the "we seem to have gotten it right" feeling), but:
// 1. it leads to incompatibities with existing data and creating the migration logic would be very expensive
// 2. there is still significant work to complete it (even without considering the costs of solving the data incompatibilty)
// So just
// 1. trim UserID back to only hold the in-domain ID
// 2. use a single domain for any given installation of the application
type UserID struct {
	IDInDomain string
	DomainID   string
}

func (userId *UserID) Equal(idInDomain, domainID string) bool {
	uid := *userId
	return idInDomain == uid.IDInDomain && domainID == uid.DomainID
}

func (userId *UserID) String() string {
	return fmt.Sprintf("%s::%s", userId.DomainID, userId.IDInDomain)
}
