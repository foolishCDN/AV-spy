package summary

type CounterOption func(*Counter)

func SetLogPrefix(prefix string) CounterOption {
	return func(c *Counter) {
		c.LogPrefix = prefix
	}
}
