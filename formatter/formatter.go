package formatter

type Formatter interface {
	Format(map[ElementName]interface{}) string
}
