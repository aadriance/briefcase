package main
import "os"
import "path"

var VERSION = "0.0.1"

type Command struct {
	invoke func()
	name   string
  description string
  usage string
}

type TempDir struct {
  path string
  envVar string
}

var commands = []Command{
  {version, "version", "Show the version of briefcase", "briefcase version"},
  {info, "info", "Show information about the temp directory used by briefcase", "briefcase info"},
  {set, "set", "Set a briefcase variable", "briefcase set <variable> <value>"},
  {get, "get", "Get a briefcase variable", "briefcase get <variable>"},
}

var envVars = []string{
  "BRIEFCASE_DIR",
  "TEMP",
  "TMPDIR",
}

func main() {
  if len(os.Args) < 2 {
    help()
    return
  }

  for _, cmd := range commands {
    if cmd.name == os.Args[1] {
      cmd.invoke()
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
  println(cmd.name)
  println("\tDescription: " + cmd.description)
  println("\tUsage: " + cmd.usage)
}

func getBriefcaseDir() string {
  return path.Join(getTempDir().path, getBriefcaseDirName())
}

// Commands to be invoked by the main program

func version() {
  println("Briefcase " + VERSION)
}

func help() {
  version()
  println("")
  for _, cmd := range commands {
    printCommandInfo(cmd)
  }
}

func info() {
  version()
  tempInfo := getTempDir()
  dirName := getBriefcaseDirName()
  println("\tTemp Dir: " + tempInfo.path)
  println("\tSourced From: " + tempInfo.envVar)
  println("\tBriefcase Directory Name: " + dirName)
}

func set() {
  briefcase := getBriefcaseDir()
  if len(os.Args) < 4 {
    println("Incorrect set usage")
    return
  }

  err := os.MkdirAll(briefcase, 0700)
  if err != nil {
    println("Error: " + err.Error())
  }

  err = os.WriteFile(path.Join(briefcase, os.Args[2] + ".txt"), []byte(os.Args[3]), 0644)
  if err != nil {
    println("Error: " + err.Error())
    return
  }
}

func get() {
  briefcase := getBriefcaseDir()
  if len(os.Args) < 3 {
    println("Incorrect get usage")
    return
  }

  data, err := os.ReadFile(path.Join(briefcase, os.Args[2] + ".txt"))
  if err != nil {
    println("Error: " + err.Error())
    return
  }

  os.Stdout.Write(data)
}
