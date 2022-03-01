package p2p

import "fmt"

func (g *networkModule) setLogger(logger func(...interface{}) (int, error)) {
	defer g.logger.Unlock()
	g.logger.Lock()

	g.logger.print = logger
}

func (g *networkModule) log(m ...interface{}) {
	defer g.logger.Unlock()
	g.logger.Lock()

	if g.logger.print != nil {
		args := make([]interface{}, 0)
		args = append(args, fmt.Sprintf("[%s]", g.address))
		args = append(args, m...)
		g.logger.print(args...)
	}
}

func (g *networkModule) clog(cond bool, m ...interface{}) {
	if cond {
		g.log(m)
	}
}
