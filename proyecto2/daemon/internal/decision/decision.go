package decision

import (
	"sort"

	"proyecto2-so1-201801391/internal/model"
)

func Rank(metrics []model.ContainerMetric) []model.ContainerMetric {
	result := append([]model.ContainerMetric(nil), metrics...)
	maxVSZ, maxRSS := uint64(1), uint64(1)
	for _, item := range result {
		if item.VSZKB > maxVSZ {
			maxVSZ = item.VSZKB
		}
		if item.RSSKB > maxRSS {
			maxRSS = item.RSSKB
		}
	}
	for index := range result {
		item := &result[index]
		vszNormalized := float64(item.VSZKB) / float64(maxVSZ) * 100
		rssNormalized := float64(item.RSSKB) / float64(maxRSS) * 100
		item.Score = item.MemoryPercent*0.35 + item.CPUPercent*0.35 +
			vszNormalized*0.15 + rssNormalized*0.15
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].Container.ID < result[j].Container.ID
		}
		return result[i].Score > result[j].Score
	})
	return result
}

func SelectVictims(metrics []model.ContainerMetric, minLow, minHigh int) []model.ContainerMetric {
	groups := map[string][]model.ContainerMetric{}
	var victims []model.ContainerMetric
	for _, metric := range metrics {
		if !metric.Container.Running {
			continue
		}
		if metric.Container.Profile == "intruder" {
			victims = append(victims, metric)
			continue
		}
		groups[metric.Container.Profile] = append(groups[metric.Container.Profile], metric)
	}
	for profile, minimum := range map[string]int{"low": minLow, "high": minHigh} {
		ranked := Rank(groups[profile])
		excess := len(ranked) - minimum
		if excess > 0 {
			victims = append(victims, ranked[:excess]...)
		}
	}
	return Rank(victims)
}
