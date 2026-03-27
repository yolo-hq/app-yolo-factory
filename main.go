package main

import (
	"github.com/yolo-hq/yolo"
	_ "github.com/yolo-hq/app-yolo-factory/server/factory/actions"
	_ "github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	_ "github.com/yolo-hq/app-yolo-factory/server/factory/filters"
	_ "github.com/yolo-hq/app-yolo-factory/server/factory/jobs"
	_ "github.com/yolo-hq/app-yolo-factory/server/factory/queries"
)

func main() {
	yolo.MustRunBinary()
}
