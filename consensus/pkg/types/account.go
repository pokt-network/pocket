package types

type Accounts struct {
	Address   string    `json:"address"`
	PublicKey PublicKey `json:"public_key"`
	UPokt     uint64    `json:"upokt"`
}

// TODO: Not implemented.
func (a *Accounts) Validate() error {
	return nil
}
