package problem1

func grayCode(n int) []int {
	totalCodes := 1 << n

	grayCodes := make([]int, 0, totalCodes)

	for i := 0; i < totalCodes; i++ {
		grayCodeValue := i ^ (i >> 1)

		grayCodes = append(grayCodes, grayCodeValue)
	}

	return grayCodes
}
