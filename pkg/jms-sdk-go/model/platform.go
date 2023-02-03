package model

type Platform struct {
	BaseOs string `json:"base"`
	ID     int    `json:"id"`
	Name   string `json:"name"`

	Protocols []PlatformProtocol `json:"protocols"`
	Category  LabelValue         `json:"category"`
	Charset   LabelValue         `json:"charset"`
	Type      LabelValue         `json:"type"`
	SuEnabled bool               `json:"su_enabled"`

	//SuMethod      string             `json:"su_method"`
	//DomainEnabled bool   `json:"domain_enabled"`
	Comment string `json:"comment"`
}

func (p *Platform) GetProtocolSetting(protocol string) (PlatformProtocol, bool) {
	for i := range p.Protocols {
		if p.Protocols[i].Name == protocol {
			return p.Protocols[i], true
		}
	}
	return PlatformProtocol{}, false
}

type PlatformProtocol struct {
	Protocol
	Setting ProtocolSetting `json:"setting"`
}
type ProtocolSetting struct {
	Security         string `json:"security"`
	Console          bool   `json:"console"`
	SftpEnabled      bool   `json:"sftp_enabled"`
	SftpHome         string `json:"sftp_home"`
	AutoFill         bool   `json:"auto_fill"`
	UsernameSelector string `json:"username_selector"`
	PasswordSelector string `json:"password_selector"`
	SubmitSelector   string `json:"submit_selector"`
}

/*
{'default': False,
 'id': 25,
 'name': 'rdp',
 'port': 3389,
 'primary': True,
 'required': False,
 'secret_types': ['password'],
 'setting': {'auto_fill': False,
			 'console': True,
			 'password_selector': '',
			 'security': 'any',
			 'sftp_enabled': True,
			 'sftp_home': '/tmp',
			 'submit_selector': '',
			 'username_selector': ''}
*/
