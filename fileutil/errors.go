package fileutil

type CfgNotExist struct {
	Message string
}

func (cne CfgNotExist) Error() string {
	return cne.Message
}

func (cne CfgNotExist) Is(target error) bool {
	_, ok := target.(CfgNotExist)

	return ok
}
