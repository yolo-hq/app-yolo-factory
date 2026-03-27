package entities

import "github.com/yolo-hq/yolo/core/registry"

func init() {
	registry.Register(Repo{}, Task{}, Run{}, Question{})
}
