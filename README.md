# Go task scheduler

## Build and run

- `go build`
- `./scheduler --jobs_path='<PATH_TO_JOBS_FILE>'`

**JOBS file example:**

```jsonc
// jobs.json
[
  {
    "name": "Job name",
    "runner": "shell",
    "path": "path_to_shell.sh", // notice the "path" key
    "repeat_every_mins": 5 // runs every 5 minutes
  },
  {
    "name": "Job name",
    "runner": "shell_inline",
    "code": "echo 'Hello' >> test_file.txt", // notice the "code" key
    "repeat_every_mins": 0.5 // runs every 30s
  }
]
```