package resolver

func New() Resolver {
	return &resolverImpl{
		version: "1.0.0",
	}
}

type resolverImpl struct {
	version string
}

func (r *resolverImpl) GetAppVersion() string {
	return "1.0.0"
}
