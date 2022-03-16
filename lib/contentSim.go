package lib

import "github.com/hbakhtiyor/strsim"

const simMultiplier = 100

// 对比响应的Body，返回一个0-100
// 的整数值，0为完全不一样，100为完
// 全一样
func ContentSim(content1, content2 string) int {

	simRate := int(strsim.Compare(content1, content2) * simMultiplier)

	return simRate
}
