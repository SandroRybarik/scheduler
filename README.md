# Go task scheduler

## Build and run

- `go build`
- `./scheduler -mode json_file_mode --jobs_path='<PATH_TO_JOBS_FILE>'`

## Modes

There are two modes of using scheduler. First mode is single json file that defines job.
Second, use _tasks/_ folder where new jobs can be added as single json files. In this mode a `"destroy": true` can be used to run a task single time after it was created.

**-mode json_file_mode**, jobs.json file example:

when using `./scheduler -mode json_file_mode -jobs_path ./jobs.json`

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


`./scheduler -mode directory_mode`, tasks example

```jsonc
// tasks/repeat_example.json
{
  "name": "Job name",
  "runner": "shell_inline",
  "code": "echo 'Hello' >> test_file.txt", // notice the "code" key
  "repeat_every_mins": 1, // runs every minute
  "destroy": false
}
```

Run just once
```jsonc
// tasks/just_once_example.json
{
  "name": "Job name",
  "runner": "shell_inline",
  "code": "echo 'Hello' >> test_file.txt", // notice the "code" key
  "repeat_every_mins": 1 // doesn't matter in this case. Job will run just once
  "destroy": true
}
```