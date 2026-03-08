package auth

import "fmt"

func fmtID(n int) string {
	return fmt.Sprintf("user_%d", n)
}
