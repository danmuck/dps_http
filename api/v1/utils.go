package v1

// currently just prepends the endpoint with a slash
func Path(endpoint string) string {
	return "/" + endpoint
}
