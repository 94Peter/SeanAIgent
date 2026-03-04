package admin

func cond(c bool, t, f string) string {
	if c {
		return t
	}
	return f
}
