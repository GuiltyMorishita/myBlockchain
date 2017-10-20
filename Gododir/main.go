package main

import do "gopkg.in/godo.v2"

func tasks(p *do.Project) {
	p.Task("server", nil, func(c *do.Context) {
		c.Start("main.go", do.M{"%in": "./"})
	}).Src("**/*.go").
		Debounce(300)

	p.Task("test", nil, func(c *do.Context) {
		c.Start("go test -v ./...", do.M{"%in": "../"})
	}).Src("**/*.go").
		Debounce(300)
}

func main() {
	do.Godo(tasks)
}
