package constants

const (
	DB_USERNAME = "amandeepbaghoria"
	DB_PASSWORD = "gm4iEWef7jywSfGc"
	SECRETKEY   = "secretkeyjwt"
	ADMIN       = "admin"
	USER        = "user"
)

func GetRole() []string {
	return []string{ADMIN, USER}
}
