package problem2

func sumOfDistancesInTree(n int, edges [][]int) []int {
	adj := make([][]int, n)
	for _, e := range edges {
		u, v := e[0], e[1]
		adj[u] = append(adj[u], v)
		adj[v] = append(adj[v], u)
	}
	subtreeSize := make([]int, n)
	ans := make([]int, n)
	var dfsCountAndSum func(node, parent int) int

	dfsCountAndSum = func(node, parent int) int {
		subtreeSize[node] = 1

		sumOfChildDist := 0

		for _, neighbor := range adj[node] {
			if neighbor != parent {
				distFromChild := dfsCountAndSum(neighbor, node)

				subtreeSize[node] += subtreeSize[neighbor]

				sumOfChildDist += distFromChild + subtreeSize[neighbor]
			}
		}

		if node == 0 {
			ans[0] = sumOfChildDist
		}

		return sumOfChildDist
	}

	dfsCountAndSum(0, -1)

	var dfsReroot func(node, parent int)

	dfsReroot = func(node, parent int) {
		if node != 0 {
			ans[node] = ans[parent] + (n - subtreeSize[node]) - subtreeSize[node]
		}

		for _, neighbor := range adj[node] {
			if neighbor != parent {
				dfsReroot(neighbor, node)
			}
		}
	}

	dfsReroot(0, -1)

	return ans
}
