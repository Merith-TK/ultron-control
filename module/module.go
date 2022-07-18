package module

// create generic module struct
type Module struct {
	Name        string
	Version     string
	Usage       string
	Description string
	Init        func()
}
