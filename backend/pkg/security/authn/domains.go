package authn

type Domain interface {
	GetDomainID() string
	CreateUserID(idInDomain string) UserID
}

type localDomain struct {
}

func (ld *localDomain) GetDomainID() string {
	return "local"
}

func (ld *localDomain) CreateUserID(idInDomain string) UserID {
	return UserID{
		DomainID:   ld.GetDomainID(),
		IDInDomain: idInDomain,
	}
}

var LocalDomain = localDomain{}
