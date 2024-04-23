package pair

import (
	"sort"
)

type pair struct {
	ID    int
	Count int
}

func GetNRelevantFromMap(data map[int]int, n int) []int {
	pairs := make([]pair, 0, len(data))
	for key, value := range data {
		pairs = append(pairs, pair{key, value})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count || (pairs[i].Count == pairs[j].Count && pairs[i].ID < pairs[j].ID)
	})
	ans := make([]int, 0, n)
	for i := 0; i < min(n, len(pairs)); i++ {
		ans = append(ans, pairs[i].ID)
	}
	return ans
}
