package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"
	"time"
)

const (
	DIRECTORY_MODE string = "directory_mode"
	JSON_FILE_MODE        = "json_file_mode"
)

type Job struct {
	Name            string  `json:"name"`
	Runner          string  `json:"runner"`
	Path            string  `json:"path"`
	Code            string  `json:"code"`
	RepeatEveryMins float64 `json:"repeat_every_mins"`
	Destroy         bool    `json:"destroy"` // delete task file after run
	lastRun         time.Time
	runAlready      bool
	path            string
}

// Load jobs from single json file.
func load(fromFile string) ([]Job, error) {
	jsonFile, err := os.Open(fromFile)

	if err != nil {
		fmt.Println(err)
		return []Job{}, nil
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var jobs []Job

	json.Unmarshal(byteValue, &jobs)

	for _, j := range jobs {
		j.runAlready = false
	}

	return jobs, nil
}

// Runs a single job and updates its data.
func runJob(job *Job) error {
	if job.Runner == "shell" {
		cmd := exec.Command("/bin/sh", job.Path)
		_, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error happened %s\n", err)
			return err
		}
		job.runAlready = true
		job.lastRun = time.Now()
		return nil
	} else if job.Runner == "shell_inline" {
		cmd := exec.Command("/bin/sh", "-c", job.Code)
		_, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error happened %s\n", err)
			return err
		}
		job.runAlready = true
		job.lastRun = time.Now()
		return nil
	}

	return errors.New("job runner: unsupported job runner type")
}

// Runs runner loop and executes tasks from jobs array.
func run(jobs []Job, done chan bool, sigs chan os.Signal) {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case sig := <-sigs:
			fmt.Println(sig)
			done <- true
		case t := <-ticker.C:
			for i := range jobs {
				// TODO: refactor me
				if jobs[i].runAlready == false {
					// run first time
					go runJob(&jobs[i])
				} else if diff := t.Sub(jobs[i].lastRun); diff.Minutes() >= float64(jobs[i].RepeatEveryMins) {
					// should it run now?
					go runJob(&jobs[i])
				}
			}
		}
	}
}

// Load tasks from dir directory and return them as jobs.
// Each task is contained in single json file in dir directory.
func loadTasksFromDirectory(dir string) ([]Job, error) {
	files, err := ioutil.ReadDir(dir)

	if err != nil {
		fmt.Println(err)
		return []Job{}, nil
	}

	var jobs []Job

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		pth := path.Join(dir, file.Name())
		fileHandle, err2 := os.Open(pth)
		defer fileHandle.Close()
		if err2 != nil {
			// handle error TODO
			continue
		}

		var job Job
		byteValue, _ := ioutil.ReadAll(fileHandle)
		json.Unmarshal(byteValue, &job)
		job.path = pth
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Runs runner loop and execute tasks from a dir folder.
// Has a same core loop as run function.
func runDir(dir string, done chan bool, sigs chan os.Signal) {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case sig := <-sigs:
			fmt.Println(sig)
			done <- true
		case t := <-ticker.C:
			// load all tasks from
			jobs, e := loadTasksFromDirectory(dir)

			if e != nil {
				fmt.Println(e)
				continue
			}

			for i := range jobs {
				// TODO: refactor me
				if jobs[i].runAlready == false {
					// run first time
					go runJob(&jobs[i])
					if jobs[i].Destroy {
						// delete task
						os.Remove(jobs[i].path)
					}
				} else if diff := t.Sub(jobs[i].lastRun); diff.Minutes() >= float64(jobs[i].RepeatEveryMins) {
					// should it run now?
					go runJob(&jobs[i])
					if jobs[i].Destroy {
						// delete task
						os.Remove(jobs[i].path)
					}
				}
			}
		}
	}
}

func main() {
	dirModeSingleTasksPath := "./tasks"
	mode := flag.String("mode", "", fmt.Sprintf("Mode: either '%s' (get tasks from single json file) or '%s' (get tasks from tasks/ folder)\n", JSON_FILE_MODE, DIRECTORY_MODE))
	jobsFilePathFlag := flag.String("jobs_path", "jobs.json", fmt.Sprintf("Jobs JSON file path when using mode=%s", JSON_FILE_MODE))

	flag.Parse()

	// Check whether json file exists if mode == JSON_FILE_MODE
	_, jobsFileErr := os.Stat(*jobsFilePathFlag)
	hasErr := *mode == JSON_FILE_MODE && errors.Is(jobsFileErr, os.ErrNotExist)

	if *mode == "" || hasErr {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)

	if *mode == DIRECTORY_MODE {
		go runDir(dirModeSingleTasksPath, done, sigs)
	} else if *mode == JSON_FILE_MODE {
		jobs, err := load(path.Join("./", *jobsFilePathFlag))

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		go run(jobs, done, sigs)
	} else {
		fmt.Println("Erorr: incorrect mode, exiting")
		os.Exit(1)
	}

	<-done
	fmt.Println("Stopping...")
}
