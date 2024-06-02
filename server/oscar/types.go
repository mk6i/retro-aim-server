package oscar

type CookieCracker interface {
	Crack(data []byte) ([]byte, error)
}
