package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
)

func main() {
	progressCmd := flag.NewFlagSet("progress", flag.ExitOnError)

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)

	lessonCmd := flag.NewFlagSet("lesson", flag.ExitOnError)
	lessonID := lessonCmd.Int("id", 1, "Lesson ID to display")

	submitCmd := flag.NewFlagSet("submit", flag.ExitOnError)
	submitLessonID := submitCmd.Int("id", 1, "Lesson ID to submit for")
	submitFile := submitCmd.String("file", "", "Path to the C source file")

	if len(os.Args) < 2 {
		fmt.Println("Usage: cli <command> [arguments]")
		fmt.Println("Commands: lesson, submit, progress, init")
		os.Exit(1)
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	cli := NewCLI(conn)

	switch os.Args[1] {
	case "progress":
		progressCmd.Parse(os.Args[2:])
		if err := cli.showProgress(); err != nil {
			log.Fatal(err)
		}

	case "init":
		initCmd.Parse(os.Args[2:])
		if err := cli.initWorkspace(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Workspace initialized successfully!")

	case "lesson":
		lessonCmd.Parse(os.Args[2:])
		if err := cli.showLesson(int32(*lessonID)); err != nil {
			log.Fatal(err)
		}

	case "submit":
		submitCmd.Parse(os.Args[2:])
		if *submitFile == "" {
			log.Fatal("must specify a file to submit")
		}
		code, err := os.ReadFile(*submitFile)
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}
		if err := cli.submitCode(int32(*submitLessonID), string(code)); err != nil {
			log.Fatal(err)
		}

	default:
		fmt.Println("Usage: cli <command> [arguments]")
		fmt.Println("Commands: lesson, submit, progress, init")
		os.Exit(1)
	}
}
