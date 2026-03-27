package main

import (
	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/registry"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"

	_ "github.com/yolo-hq/app-yolo-factory/server/factory/actions"
	_ "github.com/yolo-hq/app-yolo-factory/server/factory/filters"
	_ "github.com/yolo-hq/app-yolo-factory/server/factory/queries"
)

func main() {
	registry.Register(entities.Repo{}, entities.Task{}, entities.Run{}, entities.Question{})
	yolo.MustRunBinary()
}
