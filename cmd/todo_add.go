package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
	"github.com/spf13/cobra"
)

var todoAddCmd = &cobra.Command{
	Use:   "add [content]",
	Short: i18n.T("cmd.todo_add.short"),
	Args: func(cmd *cobra.Command, args []string) error {
		interactive, err := cmd.Flags().GetBool("interactive")
		if err != nil {
			return err
		}
		if interactive {
			return cobra.MaximumNArgs(1)(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		tasks, err := ctx.TaskStore.Load()
		if err != nil {
			return err
		}

		interactive, err := cmd.Flags().GetBool("interactive")
		if err != nil {
			return err
		}

		title := ""
		if len(args) > 0 {
			title = args[0]
		}
		parentID, err := cmd.Flags().GetInt("parent")
		if err != nil {
			return err
		}
		priority := ctx.Config.Task.DefaultPriority
		startPomo, err := cmd.Flags().GetBool("start")
		if err != nil {
			return err
		}

		if interactive {
			title, priority, parentID, startPomo, err = promptTodo(cmd.InOrStdin(), cmd.OutOrStdout(), title, priority, parentID, startPomo)
			if err != nil {
				return err
			}
		}

		task := model.Task{
			Title:     title,
			Status:    model.StatusTodo,
			Priority:  priority,
			ParentID:  nil,
			CreatedAt: time.Now(),
		}

		if parentID > 0 {
			task.ParentID = &parentID
		}

		tasks, err = ctx.TaskStore.Add(tasks, task)
		if err != nil {
			return err
		}

		task = tasks[len(tasks)-1]
		ctx.LogStore.AppendNew(model.Log{
			TodoID:    task.ID,
			Content:   task.Title,
			EventType: model.EventTaskCreate,
		})
		cmd.Println(i18n.T("cmd.todo_add.success", task.Title, task.ID))

		if startPomo {
			return ctx.RunPomodoro(cmd, task.ID)
		}

		return nil
	},
}

func init() {
	todoAddCmd.Flags().IntP("parent", "p", 0, "Parent idea ID")
	todoAddCmd.Flags().BoolP("start", "s", false, i18n.T("cmd.todo_add.flag_start"))
	todoAddCmd.Flags().BoolP("interactive", "i", false, i18n.T("cmd.todo_add.flag_interactive"))
	todoCmd.AddCommand(todoAddCmd)
}

func promptTodo(in io.Reader, out io.Writer, defaultTitle string, defaultPriority, defaultParentID int, defaultStart bool) (string, int, int, bool, error) {
	reader := bufio.NewReader(in)

	title, err := promptRequiredString(reader, out, i18n.T("cmd.todo_add.prompt_title"), defaultTitle)
	if err != nil {
		return "", 0, 0, false, err
	}

	priority, err := promptInt(reader, out, i18n.T("cmd.todo_add.prompt_priority"), defaultPriority)
	if err != nil {
		return "", 0, 0, false, err
	}

	parentID, err := promptInt(reader, out, i18n.T("cmd.todo_add.prompt_parent"), defaultParentID)
	if err != nil {
		return "", 0, 0, false, err
	}

	start, err := promptBool(reader, out, i18n.T("cmd.todo_add.prompt_start"), defaultStart)
	if err != nil {
		return "", 0, 0, false, err
	}

	return title, priority, parentID, start, nil
}

func promptRequiredString(reader *bufio.Reader, out io.Writer, label, defaultValue string) (string, error) {
	for {
		value, err := promptString(reader, out, label, defaultValue)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value), nil
		}
		fmt.Fprintln(out, i18n.T("cmd.todo_add.prompt_required"))
	}
}

func promptString(reader *bufio.Reader, out io.Writer, label, defaultValue string) (string, error) {
	if defaultValue == "" {
		fmt.Fprintf(out, "%s: ", label)
	} else {
		fmt.Fprintf(out, "%s [%s]: ", label, defaultValue)
	}
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	value = strings.TrimSpace(value)
	if err == io.EOF && value == "" && defaultValue == "" {
		return "", err
	}
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func promptInt(reader *bufio.Reader, out io.Writer, label string, defaultValue int) (int, error) {
	for {
		value, err := promptString(reader, out, label, strconv.Itoa(defaultValue))
		if err != nil {
			return 0, err
		}
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed, nil
		}
		fmt.Fprintln(out, i18n.T("cmd.todo_add.prompt_invalid_number"))
	}
}

func promptBool(reader *bufio.Reader, out io.Writer, label string, defaultValue bool) (bool, error) {
	defaultText := "n"
	if defaultValue {
		defaultText = "y"
	}
	for {
		value, err := promptString(reader, out, label, defaultText)
		if err != nil {
			return false, err
		}
		switch strings.ToLower(value) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(out, i18n.T("cmd.todo_add.prompt_invalid_bool"))
		}
	}
}
