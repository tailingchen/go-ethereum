package prometheus

type Counter interface {
	Inc()
	Add(float64)
}

type MockCounter struct{}

func (c *MockCounter) Inc()        {}
func (c *MockCounter) Add(float64) {}
