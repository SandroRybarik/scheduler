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

type Job struct {
	Name            string  `json:"name"`
	Runner          string  `json:"runner"`
	Path            string  `json:"path"`
	Code            string  `json:"code"`
	RepeatEveryMins float64 `json:"repeat_every_mins"`
	lastRun         time.Time
	runAlready      bool
}

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

func run(jobs []Job, done chan bool, sigs chan os.Signal) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case sig := <-sigs:
			fmt.Println(sig)
			done <- true
		case t := <-ticker.C:
			for i := range jobs {
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

func main() {
	// TODO: dirMode := flag.Bool("dir_mode", false, "Directory mode: get tasks from repeated_tasks/ and single_tasks/ directories")
	jobsFilePathFlag := flag.String("jobs_path", "jobs.json", "Jobs JSON file path")

	flag.Parse()

	jobs, err := load(path.Join("./", *jobsFilePathFlag))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)

	go run(jobs, done, sigs)

	<-done
	fmt.Println("Stopping...")
}
