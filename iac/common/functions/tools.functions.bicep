@description('Transforms a string key into UPPER_SNAKE_CASE format')
@export()
func toUpperSnakeCase(key string) string =>
  toUpper(replace(replace(replace(key, '-', '_'), ' ', '_'), '.', '_'))
