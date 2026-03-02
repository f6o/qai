package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/f6o/qai/internal/model"
)

type Generator struct {
	markdownDir string
}

func NewGenerator(markdownDir string) *Generator {
	return &Generator{markdownDir: markdownDir}
}

func (g *Generator) Generate(tasks []model.Task, date time.Time) (string, error) {
	var result string
	result += fmt.Sprintf("# %s (%s)\n\n", date.Format("2006-01-02"), getWeekday(date))

	ideas := g.filterIdeas(tasks)
	orphanTodos := g.filterOrphanTodos(tasks)

	sort.Slice(ideas, func(i, j int) bool {
		return ideas[i].Priority > ideas[j].Priority
	})

	for _, idea := range ideas {
		children := g.filterByParentID(tasks, idea.ID)
		result += fmt.Sprintf("## %s\n\n", idea.Title)

		if len(children) == 0 {
			result += "タスクに分解されていません。\n\n"
		} else {
			sort.Slice(children, func(i, j int) bool {
				return children[i].Priority > children[j].Priority
			})
			for _, child := range children {
				result += g.formatTask(child) + "\n"
			}
			result += "\n"
		}
	}

	if len(orphanTodos) > 0 {
		result += "## 雑多なタスク\n\n"
		sort.Slice(orphanTodos, func(i, j int) bool {
			return orphanTodos[i].Priority > orphanTodos[j].Priority
		})
		for _, t := range orphanTodos {
			result += g.formatTask(t) + "\n"
		}
		result += "\n"
	}

	return result, nil
}

func (g *Generator) Save(tasks []model.Task, date time.Time) (string, error) {
	content, err := g.Generate(tasks, date)
	if err != nil {
		return "", err
	}

	filename := filepath.Join(g.markdownDir, date.Format("2006-01-02")+".md")
	if err := os.MkdirAll(g.markdownDir, 0755); err != nil {
		return "", err
	}

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return "", err
	}

	return filename, nil
}

func (g *Generator) filterIdeas(tasks []model.Task) []model.Task {
	var result []model.Task
	for _, t := range tasks {
		if t.Status == model.StatusIdea {
			result = append(result, t)
		}
	}
	return result
}

func (g *Generator) filterOrphanTodos(tasks []model.Task) []model.Task {
	var result []model.Task
	for _, t := range tasks {
		if t.Status != model.StatusIdea && t.ParentID == nil {
			result = append(result, t)
		}
	}
	return result
}

func (g *Generator) filterByParentID(tasks []model.Task, parentID int) []model.Task {
	var result []model.Task
	for _, t := range tasks {
		if t.ParentID != nil && *t.ParentID == parentID {
			result = append(result, t)
		}
	}
	return result
}

func (g *Generator) formatTask(t model.Task) string {
	checkbox := "[ ]"
	switch t.Status {
	case model.StatusDone:
		checkbox = "[x]"
	case model.StatusDoing:
		checkbox = "[/]"
	}
	return fmt.Sprintf("- %s %s", checkbox, t.Title)
}

func getWeekday(t time.Time) string {
	weekdays := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	return weekdays[int(t.Weekday())]
}
