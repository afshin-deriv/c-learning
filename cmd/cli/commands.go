package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	pb "github.com/afshin-deriv/c-learning/proto"
	"google.golang.org/grpc"
)

type CLI struct {
	client pb.LearningServiceClient
	config Config
}

func NewCLI(conn *grpc.ClientConn) *CLI {
	cli := &CLI{
		client: pb.NewLearningServiceClient(conn),
		config: loadConfig(),
	}
	return cli
}

// initLesson creates or switches to a lesson directory and sets up the workspace
func (c *CLI) initLesson(lessonID int32) error {
	// Get lesson details
	ctx := context.Background()
	lesson, err := c.client.GetLesson(ctx, &pb.LessonRequest{
		LessonId: lessonID,
	})
	if err != nil {
		return fmt.Errorf("failed to get lesson: %v", err)
	}

	// Create lesson directory
	lessonDir := filepath.Join(c.config.WorkingDir, fmt.Sprintf("lesson%d", lessonID))
	if err := os.MkdirAll(lessonDir, 0755); err != nil {
		return fmt.Errorf("failed to create lesson directory: %v", err)
	}

	// Create initial solution.c if it doesn't exist
	solutionPath := filepath.Join(lessonDir, "solution.c")
	if _, err := os.Stat(solutionPath); os.IsNotExist(err) {
		templateCode := `#include <stdio.h>

int main() {
    // Your solution for lesson %d goes here
    return 0;
}
`
		if err := os.WriteFile(solutionPath, []byte(fmt.Sprintf(templateCode, lessonID)), 0644); err != nil {
			return fmt.Errorf("failed to create solution template: %v", err)
		}
	}

	// Create or update lesson info
	infoContent := fmt.Sprintf(`=== Lesson %d: %s ===

Description:
%s

Learning Objectives:
%s

Example Code:
%s

To complete this lesson:
1. Edit solution.c
2. Run 'cli test' to check your solution
3. Once all tests pass, you can proceed to the next lesson
`, lesson.LessonId, lesson.Title, lesson.Description,
		formatObjectives(lesson.LearningObjectives), lesson.ExampleCode)

	if err := os.WriteFile(filepath.Join(lessonDir, "README.md"), []byte(infoContent), 0644); err != nil {
		return fmt.Errorf("failed to create lesson info: %v", err)
	}

	// Create Makefile if it doesn't exist
	makefilePath := filepath.Join(lessonDir, "Makefile")
	if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
		makefile := `CC=gcc
CFLAGS=-Wall -Wextra

solution: solution.c
	$(CC) $(CFLAGS) -o $@ $<

.PHONY: clean
clean:
	rm -f solution *.o *~
`
		if err := os.WriteFile(makefilePath, []byte(makefile), 0644); err != nil {
			return fmt.Errorf("failed to create Makefile: %v", err)
		}
	}

	// Update current directory in config
	c.config.CurrentDir = lessonDir
	c.config.LastLesson = lessonID
	if err := saveConfig(c.config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Printf("Initialized Lesson %d: %s\n", lessonID, lesson.Title)
	fmt.Printf("Workspace: %s\n", lessonDir)
	fmt.Println("Edit solution.c and run 'cli test' to check your solution")
	return nil
}

// test runs the tests for the current lesson
func (c *CLI) test() error {
	// Check if we're in a lesson directory
	currentDir := c.config.CurrentDir
	if currentDir == "" {
		return fmt.Errorf("no active lesson. Run 'cli lesson --id <number>' first")
	}

	// Read the current solution
	solutionPath := filepath.Join(currentDir, "solution.c")
	code, err := os.ReadFile(solutionPath)
	if err != nil {
		return fmt.Errorf("failed to read solution: %v", err)
	}

	// Run tests
	ctx := context.Background()
	result, err := c.client.ValidateCode(ctx, &pb.CodeSubmission{
		LessonId: c.config.LastLesson,
		Code:     string(code),
	})
	if err != nil {
		return fmt.Errorf("failed to validate code: %v", err)
	}

	// Display results
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "\n=== Test Results for Lesson %d ===\n\n", c.config.LastLesson)

	allPassed := true
	for _, test := range result.TestResults {
		if !test.Passed {
			allPassed = false
		}
		status := "✓"
		if !test.Passed {
			status = "✗"
		}
		fmt.Fprintf(w, "%s\t%s\n", status, test.TestCaseDescription)
		if !test.Passed {
			fmt.Fprintf(w, "\tExpected: %s\n", test.ExpectedOutput)
			fmt.Fprintf(w, "\tGot: %s\n", test.ActualOutput)
		}
	}
	w.Flush()

	if allPassed {
		fmt.Printf("\n🎉 Congratulations! All tests passed for Lesson %d!\n", c.config.LastLesson)
		fmt.Println("You can now proceed to the next lesson with 'cli next'")
		return nil
	}

	fmt.Println("\nSome tests failed. Keep working on your solution!")
	return nil
}

// next moves to the next lesson
func (c *CLI) next() error {
	if c.config.CurrentDir == "" {
		return fmt.Errorf("no active lesson. Run 'cli lesson --id <number>' first")
	}

	nextLessonID := c.config.LastLesson + 1
	return c.initLesson(nextLessonID)
}

func formatObjectives(objectives []string) string {
	var result string
	for _, obj := range objectives {
		result += fmt.Sprintf("- %s\n", obj)
	}
	return result
}

func (c *CLI) showProgress() error {
	ctx := context.Background()
	progress, err := c.client.GetProgress(ctx, &pb.ProgressRequest{
		UserId: c.config.UserID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("\n=== Your Progress ===\n")
	fmt.Printf("Current Lesson: %d\n", progress.CurrentLesson)
	fmt.Printf("Completion: %.1f%%\n", progress.CompletionPercentage)
	fmt.Printf("Completed Lessons: %v\n", progress.CompletedLessons)
	return nil
}

func (c *CLI) initWorkspace() error {
	if err := os.MkdirAll(c.config.WorkingDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %v", err)
	}

	fmt.Printf("Initialized workspace at: %s\n", c.config.WorkingDir)
	fmt.Println("Run 'cli lesson --id 1' to start your first lesson")
	return nil
}
