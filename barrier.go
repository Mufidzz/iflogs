package iflogs

type Barrier struct {
	URL string
}

func (*Barrier) UseDefault() Engine {
	return Engine{
		Barrier: Barrier{
			URL: "localhost",
		},
	}
}
