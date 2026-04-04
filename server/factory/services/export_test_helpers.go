package services

import "github.com/yolo-hq/app-yolo-factory/server/factory/entities"

// Exported wrappers for testing from other packages.

func ExportDetermineModel(task entities.Task, project entities.Project) string {
	return determineModel(task, project)
}

func ExportWorkingDir(project entities.Project, taskID string) string {
	return workingDir(project, taskID)
}

func ExportParseTestCommands(raw string) []string {
	return parseTestCommands(raw)
}

func ExportDetectCycle(taskID string, dependsOn []string, allTasks map[string]*entities.Task) error {
	return detectCycle(taskID, dependsOn, allTasks)
}
