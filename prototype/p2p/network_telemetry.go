package p2p

import "fmt"

func (m *networkModule) setLogger(logger func(...interface{}) (int, error)) {
	defer m.logger.Unlock()
	m.logger.Lock()

	m.logger.print = logger
}

func (m *networkModule) log(args ...interface{}) {
	defer m.logger.Unlock()
	m.logger.Lock()

	if m.logger.print != nil {
		args := make([]interface{}, 0)
		args = append(args, fmt.Sprintf("[%s]", m.address))
		args = append(args, args...)
		m.logger.print(args...)
	}
}

func (m *networkModule) clog(cond bool, args ...interface{}) {
	if cond {
		m.log(args)
	}
}
