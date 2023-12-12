package model

type ConnectToken struct {
	Id       string     `json:"id"`
	User     User       `json:"user"`
	Value    string     `json:"value"`
	Account  Account    `json:"account"`
	Actions  Actions    `json:"actions"`
	Asset    Asset      `json:"asset"`
	Protocol string     `json:"protocol"`
	Domain   *Domain    `json:"domain"`
	Gateway  *Gateway   `json:"gateway"`
	ExpireAt ExpireInfo `json:"expire_at"`
	OrgId    string     `json:"org_id"`
	OrgName  string     `json:"org_name"`
	Platform Platform   `json:"platform"`

	ConnectMethod ConnectMethod `json:"connect_method"`

	ConnectOptions ConnectOptions `json:"connect_options"`

	Ticket     *ObjectId   `json:"from_ticket,omitempty"`
	TicketInfo interface{} `json:"from_ticket_info,omitempty"`

	Code   string `json:"code"`
	Detail string `json:"detail"`
	Error  string `json:"error"`
}

type ConnectMethod struct {
	Component string `json:"component"`
	Type      string `json:"type"`
	Label     string `json:"label"`
	Value     string `json:"value"`
}

type ObjectId struct {
	ID string `json:"id"`
}

type ConnectOptions struct {
	Charset          string `json:"charset"`
	DisableAutoHash  bool   `json:"disableautohash"`
	Resolution       string `json:"resolution"`
	BackspaceAsCtrlH bool   `json:"backspaceAsCtrlH"`
}
