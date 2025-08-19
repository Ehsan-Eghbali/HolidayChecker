package gate

type Result struct {
	Safe         bool
	HitCountries []string
	Errors       map[string]error
}
