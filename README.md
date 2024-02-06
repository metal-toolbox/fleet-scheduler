# Fleet Scheduler

Fleet-Scheduler is a collection of tasks that can each be set up as jobs in order to automate many things.

A **task** is a set of functions that are executed in order to make changes to fleet services.

A **job** is a task that is set up to be executed on a schedule through a [k8s cronjob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/).

## Standalone Usage

```
./fleet-scheduler <task> --config ./path/to/config.yaml
```
## Usage within sandbox

Fleet Scheduler jobs can be created within the value.yaml file of the [sandbox](https://github.com/metal-toolbox/sandbox).

Each job requires a few values in order to function.

## Creating new tasks

- Tasks need to be implemented in code within [Fleet-Scheduler](https://github.com/metal-toolbox/fleet-scheduler)
- Tasks are implemented as [cobra](https://github.com/spf13/cobra) command line commands within /cmd
- Take a look at /cmd/inventory.go for a good example.

## Creating new jobs

You just need to add the job to the [value.yaml](https://github.com/metal-toolbox/sandbox) file.
Each job requires a few values in order to function.

### Values for creating new jobs

Job Values required and location to explain them with `kubectl explain <field>`

- _name_: cronjob.metadata.name
- _restartPolicy_: cronjob.spec.jobTemplate.spec.template.spec.restartPolicy
- _imagePullPolicy_: cronjob.spec.jobTemplate.spec.template.spec.containers.imagePullPolicy
- _image_ and _tag_: cronjob.spec.jobTemplate.spec.template.spec.containers.image
- - Combined, these two values make up `cronjob.spec.jobTemplate.spec.template.spec.containers.image` like so: `${image}:${tag}`
- _ttlSecondsAfterFinished_: cronjob.spec.jobTemplate.spec.ttlSecondsafterFinished
- - This value is optional, and can be ommited
- _startingDeadlineSeconds_: cronjob.spec.startingDeadlineSeconds
- - This value is optional, and can be ommited
- _schedule_ cronjob.spec.schedule
- - Note: Does not accept cron format with second level precision. Only minute level precision
- _command_: cronjob.spec.jobTemplate.spec.template.spec.containers.command
- - This is the task to be run. Each command argument much be on a new line in array format like [so](https://stackoverflow.com/a/33136212/16289779)
- - First item in the array needs to be the binary of fleet-scheduler. Which will be `/usr/sbin/fleet-scheduler`

Example of getting details of a value with kubectl

```shell
$ kubectl explain cronjob.metadata.name
```

will give you this

```shell
KIND:       Pod
VERSION:    v1

FIELD: name <string>

DESCRIPTION:
    Name must be unique within a namespace. Is required when creating resources,
    although some resources may allow a client to request the generation of an
    appropriate name automatically. Name is primarily intended for creation
    idempotence and configuration definition. Cannot be updated. More info:
    https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names

```

## Deployment

For deployment to the [sandbox](https://github.com/metal-toolbox/sandbox), fleetscheduler.enable in value.yaml must be set to true, and the docker image must be pushed with `make push-image-devel`
