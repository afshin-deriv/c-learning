package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	pb "github.com/afshin-deriv/c-learning/proto"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedLearningServiceServer
	lessons      map[int32]*Lesson
	userProgress map[string]*UserProgress
}

type Lesson struct {
	ID                 int32      `json:"id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	ExampleCode        string     `json:"example_code"`
	LearningObjectives []string   `json:"learning_objectives"`
	TestCases          []TestCase `json:"test_cases"`
	Prerequisites      []int32    `json:"prerequisites"`
}

type LessonContent struct {
	ID                 int32    `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	LearningObjectives []string `json:"learning_objectives"`
	Prerequisites      []int32  `json:"prerequisites"`
}

type TestCase struct {
	Input       string `json:"input"`
	Expected    string `json:"expected_output"`
	Description string `json:"description"`
}

type UserProgress struct {
	CurrentLesson    int32   `json:"current_lesson"`
	CompletedLessons []int32 `json:"completed_lessons"`
}

func NewServer() *server {
	s := &server{
		lessons:      make(map[int32]*Lesson),
		userProgress: make(map[string]*UserProgress),
	}
	if err := s.loadLessons(); err != nil {
		log.Fatalf("Failed to load lessons: %v", err)
	}
	return s
}

func (s *server) loadLessons() error {
	lessonsPath := "lessons"
	return filepath.Walk(lessonsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %q: %v", path, err)
		}

		if info.IsDir() || filepath.Base(path) != "lesson.json" {
			return nil
		}

		// Read and parse lesson.json
		lessonData, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read lesson file %s: %v", path, err)
		}

		var lessonContent LessonContent
		if err := json.Unmarshal(lessonData, &lessonContent); err != nil {
			return fmt.Errorf("failed to parse lesson file %s: %v", path, err)
		}

		// Read example code
		exampleCode, err := os.ReadFile(filepath.Join(filepath.Dir(path), "example.c"))
		if err != nil {
			return fmt.Errorf("failed to read example code for lesson %d: %v", lessonContent.ID, err)
		}

		// Read test cases
		testData, err := os.ReadFile(filepath.Join(filepath.Dir(path), "tests.json"))
		if err != nil {
			return fmt.Errorf("failed to read test cases for lesson %d: %v", lessonContent.ID, err)
		}

		var testCases []TestCase
		if err := json.Unmarshal(testData, &testCases); err != nil {
			return fmt.Errorf("failed to parse test cases for lesson %d: %v", lessonContent.ID, err)
		}

		// Create complete lesson
		lesson := &Lesson{
			ID:                 lessonContent.ID,
			Title:              lessonContent.Title,
			Description:        lessonContent.Description,
			ExampleCode:        string(exampleCode),
			LearningObjectives: lessonContent.LearningObjectives,
			TestCases:          testCases,
			Prerequisites:      lessonContent.Prerequisites,
		}

		s.lessons[lesson.ID] = lesson
		log.Printf("Loaded lesson %d: %s", lesson.ID, lesson.Title)
		return nil
	})
}

func (s *server) GetLesson(ctx context.Context, req *pb.LessonRequest) (*pb.LessonResponse, error) {
	lesson, ok := s.lessons[req.LessonId]
	if !ok {
		return nil, fmt.Errorf("lesson %d not found", req.LessonId)
	}

	return &pb.LessonResponse{
		LessonId:           lesson.ID,
		Title:              lesson.Title,
		Description:        lesson.Description,
		ExampleCode:        lesson.ExampleCode,
		LearningObjectives: lesson.LearningObjectives,
		TestCases:          convertTestCases(lesson.TestCases),
	}, nil
}

func (s *server) ValidateCode(ctx context.Context, req *pb.CodeSubmission) (*pb.ValidationResponse, error) {
	lesson, ok := s.lessons[req.LessonId]
	if !ok {
		return nil, fmt.Errorf("lesson %d not found", req.LessonId)
	}

	// Create temporary directory for compilation
	tmpDir, err := os.MkdirTemp("", "c-learning-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	results, err := s.compileAndRunTests(req.Code, lesson.TestCases, tmpDir)
	if err != nil {
		return &pb.ValidationResponse{
			IsValid:  false,
			Feedback: err.Error(),
		}, nil
	}

	allPassed := true
	for _, result := range results {
		if !result.Passed {
			allPassed = false
			break
		}
	}

	return &pb.ValidationResponse{
		IsValid:     allPassed,
		TestResults: results,
		Feedback:    getFeedback(allPassed, results),
		CanProceed:  allPassed,
	}, nil
}

func (s *server) GetProgress(ctx context.Context, req *pb.ProgressRequest) (*pb.ProgressResponse, error) {
	progress, ok := s.userProgress[req.UserId]
	if !ok {
		progress = &UserProgress{
			CurrentLesson:    1,
			CompletedLessons: []int32{},
		}
		s.userProgress[req.UserId] = progress
	}

	totalLessons := len(s.lessons)
	completionPercentage := float32(len(progress.CompletedLessons)) / float32(totalLessons) * 100

	return &pb.ProgressResponse{
		CurrentLesson:        progress.CurrentLesson,
		CompletedLessons:     progress.CompletedLessons,
		CompletionPercentage: completionPercentage,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterLearningServiceServer(s, NewServer())

	log.Printf("Server listening on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
