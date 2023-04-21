package kvstore

// PostgresKV implements the KVStore interface.
type PostgresKV struct{}

// Get ...
func (p *PostgresKV) Get(key []byte) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

// Set ...
func (p *PostgresKV) Set(key []byte, value []byte) error {
	panic("not implemented") // TODO: Implement
}

// Delete ...
func (p *PostgresKV) Delete(key []byte) error {
	panic("not implemented") // TODO: Implement
}

// Lifecycle methods

// Stop ...
func (p *PostgresKV) Stop() error {
	panic("not implemented") // TODO: Implement
}

// Accessors

// GetAll ...
func (p *PostgresKV) GetAll(prefixKey []byte, descending bool) (keys [][]byte, values [][]byte, err error) {
	panic("not implemented") // TODO: Implement
}

// Exists ...
func (p *PostgresKV) Exists(key []byte) (bool, error) {
	panic("not implemented") // TODO: Implement
}

// ClearAll ...
func (p *PostgresKV) ClearAll() error {
	panic("not implemented") // TODO: Implement
}
