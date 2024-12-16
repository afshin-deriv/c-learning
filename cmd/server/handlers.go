package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	pb "github.com/afshin-deriv/c-learning/proto"
)

// convertTestCases converts internal TestCase format to protobuf format
func convertTestCases(cases []TestCase) []*pb.TestCase {
	result := make([]*pb.TestCase, len(cases))
	for i, tc := range cases {
		result[i] = &pb.TestCase{
			Input:          tc.Input,
			ExpectedOutput: tc.Expected,
			Description:    tc.Description,
		}
	}
	return result
}

// getFeedback generates feedback message based on test results
func getFeedback(allPassed bool, results []*pb.TestResult) string {
	if allPassed {
		return "Great job! All tests passed successfully."
	}

	var failedCount int
	for _, result := range results {
		if !result.Passed {
			failedCount++
		}
	}

	if failedCount == len(results) {
		return "None of the tests passed. Review your code and try again."
	}

	return fmt.Sprintf("%d out of %d tests failed. Check the test results and try again.",
		failedCount, len(results))
}

// validatePrerequisites checks if user has completed required prerequisites
func (s *server) validatePrerequisites(lessonID int32, progress *UserProgress) bool {
	lesson, ok := s.lessons[lessonID]
	if !ok {
		return false
	}

	for _, prereq := range lesson.Prerequisites {
		found := false
		for _, completed := range progress.CompletedLessons {
			if completed == prereq {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// updateProgress updates user progress after successful completion
func (s *server) updateProgress(userID string, lessonID int32) {
	progress, ok := s.userProgress[userID]
	if !ok {
		progress = &UserProgress{
			CurrentLesson:    1,
			CompletedLessons: []int32{},
		}
	}

	// Check if lesson is already completed
	for _, completed := range progress.CompletedLessons {
		if completed == lessonID {
			return
		}
	}

	// Add to completed lessons
	progress.CompletedLessons = append(progress.CompletedLessons, lessonID)

	// Update current lesson if this was the current one
	if progress.CurrentLesson == lessonID {
		progress.CurrentLesson = lessonID + 1
	}

	s.userProgress[userID] = progress
}

// getNextAvailableLesson finds the next lesson user can take
func (s *server) getNextAvailableLesson(userID string) int32 {
	progress, ok := s.userProgress[userID]
	if !ok {
		return 1
	}

	current := progress.CurrentLesson
	for {
		if _, ok := s.lessons[current]; !ok {
			return progress.CurrentLesson
		}

		if s.validatePrerequisites(current, progress) {
			return current
		}

		current++
	}
}

// compileAndRunTests handles code compilation and test execution
func (s *server) compileAndRunTests(code string, testCases []TestCase, tmpDir string) ([]*pb.TestResult, error) {
	srcFile := filepath.Join(tmpDir, "solution.c")
	if err := os.WriteFile(srcFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write source file: %v", err)
	}

	outFile := filepath.Join(tmpDir, "solution")
	cmd := exec.Command("gcc", "-o", outFile, srcFile, "-Wall", "-Werror")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("compilation failed:\n%s", string(output))
	}

	var results []*pb.TestResult
	for _, tc := range testCases {
		cmd := exec.Command(outFile)
		cmd.Stdin = strings.NewReader(tc.Input)
		output, err := cmd.CombinedOutput()

		passed := err == nil && strings.TrimSpace(string(output)) == strings.TrimSpace(tc.Expected)
		results = append(results, &pb.TestResult{
			Passed:              passed,
			TestCaseDescription: tc.Description,
			ActualOutput:        string(output),
			ExpectedOutput:      tc.Expected,
		})
	}

	return results, nil
}
