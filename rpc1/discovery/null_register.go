package discovery

type NullRegiser struct {
}

func NewNullRegister() *NullRegiser {
	return &NullRegiser{}
}

func (e *NullRegiser) Regist(name, host string, port int) {
}

func (e *NullRegiser) Revoke() {
}
