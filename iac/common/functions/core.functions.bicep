@description('Builds the name of a resource')
@export()
func BuildResourceName(prefix string, trigram string, number string) string =>
  '${prefix}${trigram}${number}'
