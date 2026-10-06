package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func main() {

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}

type Task struct {
	ID          int
	Description string
	CreatedAt   time.Time
	isCompleted bool
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Run: func(cmd *cobra.Command, args []string) {
		tasks, err := LoadTask("tasks.csv")
		if err != nil {
			fmt.Println("Error loading tasks:", err)
			return
		}
		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return
		}

		for _, task := range tasks {
			status := "Pending"
			if task.isCompleted {
				status = "Completed"
			}
			fmt.Printf("ID: %d, Description: %s, Status: %s\n", task.ID, task.Description, status)
		}
	},
}

var completeCmd = &cobra.Command{
	Use:   "complete [task ID]",
	Short: "Mark a task as completed",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid task ID")
			return
		}

		tasks, err := LoadTask("tasks.csv")
		if err != nil {
			fmt.Println("Error loading tasks:", err)
			return
		}

		found := false
		for i := range tasks {
			if tasks[i].ID == taskID {
				found = true
				if tasks[i].isCompleted {
					fmt.Println("Task is already marked as completed.")
					return
				}

				tasks[i].isCompleted = true
				break
			}
		}

		if !found {
			fmt.Println("Task not found.")
			return
		}

		if err := saveTask(tasks, "tasks.csv"); err != nil {
			fmt.Println("Error saving task:", err)
			return
		}

		fmt.Println("Task", taskID, "marked as completed.")

	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [task id]",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid task ID")
			return
		}
		loadedTasks, err := LoadTask("tasks.csv")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error deleting task:", err)
			os.Exit(1)
		}

		var remainingTasks []Task

		found := false
		for _, task := range loadedTasks {
			if task.ID == taskID {
				found = true
				continue
			}
			remainingTasks = append(remainingTasks, task)
		}
		if !found {
			fmt.Println(taskID, "Task not found.")
			os.Exit(1)
		}
		if err := saveTask(remainingTasks, "tasks.csv"); err != nil {
			fmt.Fprintln(os.Stderr, "Error updating task:", err)
			os.Exit(1)
		}
		fmt.Println("Task", taskID, "deleted successfully.")
	},
}

var rootCmd = &cobra.Command{
	Use:   "tasks",
	Short: "A CLI task manager",
}

var addCmd = &cobra.Command{
	Use:   "add [description]",
	Short: "Add a new task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		existingTasks, err := LoadTask("tasks.csv")
		if err != nil {
			fmt.Println("Error loading tasks:", err)
			os.Exit(1)
			return
		}

		newTask := Task{
			ID:          rand.Intn(100000),
			Description: args[0],
			CreatedAt:   time.Now(),
			isCompleted: false,
		}
		existingTasks = append(existingTasks, newTask)

		if err := saveTask(existingTasks, "tasks.csv"); err != nil {
			fmt.Println("Error saving task:", err)
			os.Exit(1)
			return

		}
		fmt.Println("Task added successfully!", newTask.Description)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(completeCmd)
	rootCmd.AddCommand(deleteCmd)
}
func saveTask(tasks []Task, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	WriteHeader := []string{"ID", "Description", "CreatedAt", "isCompleted"}
	writer.Write(WriteHeader)
	if err := writer.Error(); err != nil {
		return err

	}
	for _, task := range tasks {
		row := []string{
			strconv.Itoa(task.ID),
			task.Description,
			task.CreatedAt.Format(time.RFC3339),
			strconv.FormatBool(task.isCompleted),
		}
		writer.Write(row)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}

	return nil

}

func LoadTask(filepath string) ([]Task, error) {

	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}, nil
		}
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return []Task{}, nil
	}
	var tasks []Task
	for _, record := range records[1:] { // Skip header row
		recordID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, err
		}
		Description := record[1]
		if Description == "" {
			return nil, fmt.Errorf("invalid task description")
		}

		recordCreatedAt, err := time.Parse(time.RFC3339, record[2])
		if err != nil {
			return nil, err
		}
		recordIsCompleted, err := strconv.ParseBool(record[3])
		if err != nil {
			return nil, err
		}
		task := Task{
			ID:          recordID,
			Description: Description,
			CreatedAt:   recordCreatedAt,
			isCompleted: recordIsCompleted,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil

}
