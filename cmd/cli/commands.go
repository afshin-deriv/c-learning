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

func (c *CLI) showLesson(lessonID int32) error {
	ctx := context.Background()
	lesson, err := c.client.GetLesson(ctx, &pb.LessonRequest{
		LessonId: lessonID,
	})
	if err != nil {
		return fmt.Errorf("failed to get lesson: %v", err)
	}

	fmt.Printf("\n=== Lesson %d: %s ===\n\n", lesson.LessonId, lesson.Title)
	fmt.Printf("Description:\n%s\n\n", lesson.Description)

	fmt.Println("Learning Objectives:")
	for _, obj := range lesson.LearningObjectives {
		fmt.Printf("- %s\n", obj)
	}

	fmt.Printf("\nExample Code:\n%s\n", lesson.ExampleCode)
	return nil
}

func (c *CLI) submitCode(lessonID int32, code string) error {
	ctx := context.Background()
	result, err := c.client.ValidateCode(ctx, &pb.CodeSubmission{
		LessonId: lessonID,
		Code:     code,
	})
	if err != nil {
		return fmt.Errorf("failed to validate code: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "\n=== Test Results ===\n\n")
	for _, test := range result.TestResults {
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

	if result.CanProceed {
		fmt.Printf("\n🎉 Congratulations! You've completed lesson %d!\n", lessonID)
		c.config.LastLesson = lessonID + 1
		if err := saveConfig(c.config); err != nil {
			return fmt.Errorf("failed to save progress: %v", err)
		}
	}

	return nil
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

	// Create template.c
	templateC := `#include <stdio.h>

int main() {
    // Your code here
    return 0;
}
`
	if err := os.WriteFile(filepath.Join(c.config.WorkingDir, "template.c"), []byte(templateC), 0644); err != nil {
		return fmt.Errorf("failed to create template: %v", err)
	}

	// Create Makefile
	makefile := `CC=gcc
CFLAGS=-Wall -Wextra

%: %.c
	$(CC) $(CFLAGS) -o $@ $<

.PHONY: clean
clean:
	rm -f *.o *~
`
	if err := os.WriteFile(filepath.Join(c.config.WorkingDir, "Makefile"), []byte(makefile), 0644); err != nil {
		return fmt.Errorf("failed to create Makefile: %v", err)
	}

	fmt.Printf("Initialized workspace at: %s\n", c.config.WorkingDir)
	return nil
}
