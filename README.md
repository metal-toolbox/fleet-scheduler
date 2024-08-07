# Fleet Scheduler

Fleet-Scheduler is a collection of tasks that can each be set up as jobs in order to automate many things.

A **task** is a set of functions that are executed in order to make changes to fleet services.

A **job** is a task that is set up to be executed on a schedule through a [k8s cronjob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/).

## Standalone Usage

```
./fleet-scheduler <task> --config ./path/to/config.yaml
```
## Usage within sandbox

Sandbox will include the [helm chart](https://github.com/metal-toolbox/fleet-scheduler/tree/main/chart) by including it as a dependency in Chart.yaml

Fleet Scheduler jobs can be created within the value.yaml.

Each job requires a few values in order to function.

## Creating new tasks

- Tasks need to be implemented in code within [Fleet-Scheduler](https://github.com/metal-toolbox/fleet-scheduler)
- Tasks are implemented as [cobra](https://github.com/spf13/cobra) command line commands within /cmd
- Take a look at /cmd/inventory.go for a good example.

## Creating new jobs

Jobs can be created within the Sandbox, or within [values.yaml](https://github.com/metal-toolbox/fleet-scheduler/tree/main/chart/values.yaml) here.

Explained in the sandbox [README](https://github.com/metal-toolbox/sandbox/blob/main/notes/fleet-scheduler.md).
