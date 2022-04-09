package monitor

import "sync"

var ProgressBars  = map[string]ProgressBar{}

type ProgressBar struct {
	Total int
	Progress *sync.Map
}

type ProgressBarJson struct {
	Name string
	Total int
	Progress int
	Status int  // 101 运行中  200 执行完成
}