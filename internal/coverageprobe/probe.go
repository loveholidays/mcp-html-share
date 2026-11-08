package coverageprobe

func Increment(value int) int {
	if value < 0 {
		return 0
	}
	return value + 1
}
