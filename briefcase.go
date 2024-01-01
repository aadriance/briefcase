package main

import (
	"fmt"
	"os"
	"path"
	"strings"
)

var VERSION = "0.0.1"

type Command struct {
	invoke      func(UserArgs)
	name        string
	description string
	usage       string
}

type TempDir struct {
	path   string
	envVar string
}

type UserArgs struct {
	name  string
	value string
}

var commands = []Command{
	{version, "version", "Show the version of briefcase", "briefcase version"},
	{info, "info", "Show information about the temp directory used by briefcase", "briefcase info"},
	{set, "set", "Set a briefcase variable", "briefcase set <variable> <value>"},
	{get, "get", "Get a briefcase variable", "briefcase get <variable>"},
	{purge, "purge", "Purge briefcase data", "briefcase purge"},
	{remove, "remove", "Remove a briefcase variable", "briefcase remove <variable>"},
	{list, "list", "List briefcase entries", "briefcase list"},
}

var envVars = []string{
	"BRIEFCASE_DIR",
	"TEMP",
	"TMPDIR",
}

func main() {
	args := UserArgs{"", ""}

	if len(os.Args) < 2 {
		help()
		return
	}

	if len(os.Args) > 2 {
		args.name = os.Args[2]
	}

	if len(os.Args) > 3 {
		args.value = strings.Join(os.Args[3:len(os.Args)], " ")
	}

	for _, cmd := range commands {
		if cmd.name == os.Args[1] {
			cmd.invoke(args)
			return
		}
	}

	help()
}

// utility functions

func getTempDir() TempDir {
	for _, envVar := range envVars {
		if dir := os.Getenv(envVar); dir != "" {
			return TempDir{dir, envVar}
		}
	}

	return TempDir{"/tmp", "N/A"}
}

func getBriefcaseDirName() string {
	if name := os.Getenv("BRIEFCASE_DIRNAME"); name != "" {
		return name
	}

	return "briefcase"
}

func printCommandInfo(cmd Command) {
	fmt.Println(cmd.name)
	fmt.Println("\tDescription: " + cmd.description)
	fmt.Println("\tUsage: " + cmd.usage)
}

func getBriefcaseDir() string {
	return path.Join(getTempDir().path, getBriefcaseDirName())
}

// Commands to be invoked by the main program

func version(_ UserArgs) {
	fmt.Println("Briefcase " + VERSION)
}

func help() {
	fmt.Println("")
	for _, cmd := range commands {
		printCommandInfo(cmd)
	}
}

func info(_ UserArgs) {
	tempInfo := getTempDir()
	dirName := getBriefcaseDirName()
	fmt.Println("\tTemp Dir: " + tempInfo.path)
	fmt.Println("\tSourced From: " + tempInfo.envVar)
	fmt.Println("\tBriefcase Directory Name: " + dirName)
}

func set(args UserArgs) {
	briefcase := getBriefcaseDir()
	if args.name == "" || args.value == "" {
		fmt.Println("Incorrect set usage")
		return
	}

	err := os.MkdirAll(briefcase, 0700)
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}

	err = os.WriteFile(path.Join(briefcase, args.name), []byte(args.value), 0644)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return
	}
}

func get(args UserArgs) {
	briefcase := getBriefcaseDir()
	if args.name == "" {
		fmt.Println("Incorrect get usage")
		return
	}

	data, err := os.ReadFile(path.Join(briefcase, args.name))
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return
	}

	os.Stdout.Write(data)
}

func purge(args UserArgs) {
	var confirm string
	if args.name == "force" {
		confirm = "y"
	} else {
		fmt.Println("Are you sure you want to delete all briefcase data? (y/n)")
		fmt.Scan(&confirm)
	}

	if confirm != "y" {
		fmt.Println("Exiting without deleting data")
	} else {
		briefcase := getBriefcaseDir()
		os.RemoveAll(briefcase)
	}
}

func remove(args UserArgs) {
	briefcase := getBriefcaseDir()
	if args.name == "" {
		fmt.Println("Incorrect remove usage")
		return
	}

	err := os.Remove(path.Join(briefcase, args.name))
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}
}

func list(_ UserArgs) {
	briefcase := getBriefcaseDir()
	files, err := os.ReadDir(briefcase)
	if err != nil {
		// If there's an error, it's because the briefcase dir doesn't exist.
		// simply list nothing.
		return
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}
}
