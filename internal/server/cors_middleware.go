package server

var allowedOrigins = []string{
	"http://localhost:9009",
	"https://your-production-site.com",
}

func isAllowedOrigin(origin string) bool {
	for _, o := range allowedOrigins {
		if origin == o {
			return true
		}
	}
	return false
}
