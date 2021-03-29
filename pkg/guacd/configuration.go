package guacd

func NewConfiguration() (ret Configuration) {
	ret.Parameters = make(map[string]string)
	return ret
}

type Configuration struct {
	ConnectionID string
	Protocol     string
	Parameters   map[string]string
}

func (opt *Configuration) SetParameter(name, value string) {
	opt.Parameters[name] = value
}

func (opt *Configuration) UnSetParameter(name string) {
	delete(opt.Parameters, name)
}

func (opt *Configuration) GetParameter(name string) string {
	return opt.Parameters[name]
}
