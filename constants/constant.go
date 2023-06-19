package constants

const (
	DB_USERNAME = "amandeepbaghoria"
	DB_PASSWORD = "gm4iEWef7jywSfGc"
	SECRETKEY   = "secretkeyjwt"
)

func GetRole() []string {
	return []string{"admin", "user"}
}
