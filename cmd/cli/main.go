package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
)

func main() {
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	nextCmd := flag.NewFlagSet("next", flag.ExitOnError)
	progressCmd := flag.NewFlagSet("progress", flag.ExitOnError)
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)

	lessonCmd := flag.NewFlagSet("lesson", flag.ExitOnError)
	lessonID := lessonCmd.Int("id", 1, "Lesson ID to start")

	if len(os.Args) < 2 {
		fmt.Println("Usage: cli <command> [arguments]")
		fmt.Println("Commands: lesson, test, next, progress, init")
		os.Exit(1)
	}

	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	cli := NewCLI(conn)

	switch os.Args[1] {
	case "lesson":
		lessonCmd.Parse(os.Args[2:])
		if err := cli.initLesson(int32(*lessonID)); err != nil {
			log.Fatal(err)
		}

	case "test":
		testCmd.Parse(os.Args[2:])
		if err := cli.test(); err != nil {
			log.Fatal(err)
		}

	case "next":
		nextCmd.Parse(os.Args[2:])
		if err := cli.next(); err != nil {
			log.Fatal(err)
		}

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

	default:
		fmt.Println("Usage: cli <command> [arguments]")
		fmt.Println("Commands: lesson, test, next, progress, init")
		os.Exit(1)
	}
}
