package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/redisq"
)

//go:generate mockgen -source drivly_task_service.go -destination mocks/drivly_task_service_mock.go
type DrivlyTaskService interface {
	StartDrivlyUpdate(deviceDefinitionID, userDeviceID, vin string) (taskID string, err error)
	GetTaskStatus(ctx context.Context, taskID string) (task *DrivlyTask, err error)
	StartConsumer(ctx context.Context)
}

// task names
const (
	updateDrivlyTask = "updateDrivlyTask"
)

func NewDrivlyTaskService(settings *config.Settings, deviceDefinitionSvc DeviceDefinitionService, logger zerolog.Logger) DrivlyTaskService {
	// setup redis connection
	var tlsConfig *tls.Config
	if settings.RedisTLS {
		tlsConfig = new(tls.Config)
	}
	var r StandardRedis
	// handle redis cluster in prod
	if settings.Environment == "prod" {
		r = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:     []string{settings.RedisURL},
			Password:  settings.RedisPassword,
			TLSConfig: tlsConfig,
		})
	} else {
		r = redis.NewClient(&redis.Options{
			Addr:      settings.RedisURL,
			Password:  settings.RedisPassword,
			TLSConfig: tlsConfig,
		})
	}

	var QueueFactory = redisq.NewFactory()
	const workerQueueName = "drivly-worker"
	// Create a queue.
	mainQueue := QueueFactory.RegisterQueue(&taskq.QueueOptions{
		Name:  workerQueueName,
		Redis: r, // go-redis client
	})
	// register task, handler would be below as
	dls := &drivlyTaskService{
		settings:            settings,
		mainQueue:           mainQueue,
		deviceDefinitionSvc: deviceDefinitionSvc,
		log:                 logger.With().Str("worker queue", workerQueueName).Logger(),
	}
	updateTask := taskq.RegisterTask(&taskq.TaskOptions{
		Name: updateDrivlyTask,
		Handler: func(ctx context.Context, taskID, deviceID, userID, unitID string) error {
			return dls.ProcessUpdate(ctx, taskID, deviceID, userID, unitID)
		},
		RetryLimit: 5,
		MinBackoff: time.Second * 30,
		MaxBackoff: time.Minute,
	})

	dls.updateDrivlyTask = updateTask
	dls.redis = r
	return dls
}

type drivlyTaskService struct {
	settings            *config.Settings
	mainQueue           taskq.Queue
	updateDrivlyTask    *taskq.Task
	redis               StandardRedis
	deviceDefinitionSvc DeviceDefinitionService
	log                 zerolog.Logger
}

// StartDrivlyUpdate produces a task to pull vin data
func (dls *drivlyTaskService) StartDrivlyUpdate(deviceDefinitionID, userDeviceID, vin string) (taskID string, err error) {
	taskID = ksuid.New().String()
	msg := dls.updateDrivlyTask.WithArgs(context.Background(), taskID, deviceDefinitionID, userDeviceID, vin)
	msg.Name = taskID
	err = dls.mainQueue.Add(msg)
	if err != nil {
		return "", err
	}
	err = dls.updateTaskState(taskID, "waiting for task to be processed", Pending, 100, nil)
	if err != nil {
		return taskID, err
	}

	return taskID, nil
}

func (dls *drivlyTaskService) StartConsumer(ctx context.Context) {
	if err := dls.mainQueue.Consumer().Start(ctx); err != nil {
		dls.log.Err(err).Msg("consumer failed")
	}
	dls.log.Info().Msg("started drivly tasks consumer")
}

// GetTaskStatus gets the status from the redis backend - is there a way to do this? multistep
func (dls *drivlyTaskService) GetTaskStatus(ctx context.Context, taskID string) (task *DrivlyTask, err error) {
	// problem is taskq does not have a way to retrieve a task, and we want to persist state as we move along the task
	taskRaw := dls.redis.Get(ctx, buildDrivlyTaskRedisKey(taskID))
	if taskRaw == nil {
		return nil, errors.New("task not found")
	}
	taskBytes, err := taskRaw.Bytes()
	if err != nil {
		return nil, err
	}
	drivlyTask := new(DrivlyTask)
	err = json.Unmarshal(taskBytes, drivlyTask)
	if err != nil {
		return nil, err
	}
	return drivlyTask, nil
}

// ProcessUpdate handles the drivly update task when consumed
func (dls *drivlyTaskService) ProcessUpdate(ctx context.Context, taskID, deviceDefinitionID, userDeviceID, vin string) error {
	// check for ctx.Done in channel etc but in non-blocking way? to then return err if so to retry on next app reboot

	log := dls.log.With().Str("handler", updateDrivlyTask+"_ProcessUpdate").
		Str("taskID", taskID).
		Str("userDeviceID", userDeviceID).
		Str("vin", vin).
		Str("deviceDefinitionID", deviceDefinitionID).Logger()
	// store the userID?
	log.Info().Msg("Started processing autopi update")

	err := dls.updateTaskState(taskID, "Started", InProcess, 110, nil)
	if err != nil {
		log.Err(err).Msg("failed to persist state to redis")
		return err
	}

	_, err = dls.deviceDefinitionSvc.PullDrivlyData(ctx, userDeviceID, deviceDefinitionID, vin, false)

	if err != nil {
		log.Err(err).Msg("Fail to update information")
		return err
	}

	_ = dls.updateTaskState(taskID, "autopi update confirmed", Success, 200, nil)
	return nil // by not returning error, task will not be processed again.
}

func (dls *drivlyTaskService) updateTaskState(taskID, description string, status TaskStatusEnum, code int, err error) error {
	updateCnt := 0
	existing, _ := dls.GetTaskStatus(context.Background(), taskID)
	if existing != nil {
		updateCnt = existing.Updates + 1
	}
	t := DrivlyTask{
		TaskID:      taskID,
		Status:      string(status),
		Description: description,
		Code:        code, // need enum with codes
		UpdatedAt:   time.Now().UTC(),
		Updates:     updateCnt,
	}
	if err != nil {
		errstr := err.Error()
		t.Error = &errstr
	}
	jb, errM := json.Marshal(t)
	if errM != nil {
		return errM
	}
	set := dls.redis.Set(context.Background(), buildDrivlyTaskRedisKey(taskID), jb, time.Hour*72)
	return set.Err()
}

func buildDrivlyTaskRedisKey(taskID string) string {
	return updateDrivlyTask + "_" + taskID
}

// DrivlyTask describes a task that is being worked on asynchronously for autopi
type DrivlyTask struct {
	TaskID      string    `json:"taskId"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Code        int       `json:"code"`
	Error       *string   `json:"error,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
	// Updates increments every time the job was updated.
	Updates int `json:"updates"`
}
