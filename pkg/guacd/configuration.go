package guacd

func NewConfiguration() (conf Configuration) {
	conf.Parameters = make(map[string]string)
	return conf
}

type Configuration struct {
	ConnectionID string
	Protocol     string
	Parameters   map[string]string
}

func (conf *Configuration) SetParameter(name, value string) {
	conf.Parameters[name] = value
}

func (conf *Configuration) UnSetParameter(name string) {
	delete(conf.Parameters, name)
}

func (conf *Configuration) GetParameter(name string) string {
	return conf.Parameters[name]
}
