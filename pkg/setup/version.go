package setup

func getMajorMinorVersion(s string) string {
	dot := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			dot++
		}
		if dot == 2 {
			return s[:i]
		}
	}
	return s
}
