package services

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/redisq"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

//go:generate mockgen -source autopi_task_service.go -destination mocks/autopi_task_service_mock.go
type AutoPiTaskService interface {
	StartQueryAndUpdateVIN(deviceID, unitID, userDeviceID string) (taskID string, err error)
	GetTaskStatus(ctx context.Context, taskID string) (task *AutoPiTask, err error)
	StartConsumer(ctx context.Context)
}

// task names
const (
	updateAutoPiTask      = "updateTask"
	queryAndUpdateVINTask = "queryAndUpdateVINTask"
)

func NewAutoPiTaskService(settings *config.Settings, autoPiSvc AutoPiAPIService, dbs func() *db.ReaderWriter, logger zerolog.Logger) AutoPiTaskService {
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
	const workerQueueName = "autopi-worker"
	// Create a queue.
	mainQueue := QueueFactory.RegisterQueue(&taskq.QueueOptions{
		Name:  workerQueueName,
		Redis: r, // go-redis client
	})
	// register task, handler would be below as
	ats := &autoPiTaskService{
		settings:  settings,
		mainQueue: mainQueue,
		autoPiSvc: autoPiSvc,
		dbs:       dbs,
		log:       logger.With().Str("worker queue", workerQueueName).Logger(),
	}
	vinTask := taskq.RegisterTask(&taskq.TaskOptions{
		Name: queryAndUpdateVINTask,
		Handler: func(ctx context.Context, taskID, deviceID, unitID, userDeviceID string) error {
			return ats.queryAndUpdateVIN(ctx, taskID, deviceID, unitID, userDeviceID)
		},
		RetryLimit: 1,
		MinBackoff: time.Second * 5,
		MaxBackoff: time.Second * 30,
	})

	ats.getAndSetVinTask = vinTask
	ats.redis = r
	return ats
}

type autoPiTaskService struct {
	settings         *config.Settings
	mainQueue        taskq.Queue
	getAndSetVinTask *taskq.Task
	redis            StandardRedis
	autoPiSvc        AutoPiAPIService
	dbs              func() *db.ReaderWriter
	log              zerolog.Logger
}

func (ats *autoPiTaskService) StartQueryAndUpdateVIN(deviceID, userID, unitID string) (taskID string, err error) {
	taskID = ksuid.New().String()
	msg := ats.getAndSetVinTask.WithArgs(context.Background(), deviceID, userID, unitID)
	msg.Name = taskID
	err = ats.mainQueue.Add(msg)

	if err != nil {
		return "", err
	}
	err = ats.updateTaskState(taskID, "waiting for task to be processed", Pending, 100, nil)
	if err != nil {
		return taskID, err
	}

	return taskID, nil
}

func (ats *autoPiTaskService) StartConsumer(ctx context.Context) {
	if err := ats.mainQueue.Consumer().Start(ctx); err != nil {
		ats.log.Err(err).Msg("consumer failed")
	}
	ats.log.Info().Msg("started autopi tasks consumer")
}

// GetTaskStatus gets the status from the redis backend - is there a way to do this? multistep
func (ats *autoPiTaskService) GetTaskStatus(ctx context.Context, taskID string) (task *AutoPiTask, err error) {
	// problem is taskq does not have a way to retrieve a task, and we want to persist state as we move along the task
	taskRaw := ats.redis.Get(ctx, buildAutoPiTaskRedisKey(taskID))
	if taskRaw == nil {
		return nil, errors.New("task not found")
	}
	taskBytes, err := taskRaw.Bytes()
	if err != nil {
		return nil, err
	}
	apTask := new(AutoPiTask)
	err = json.Unmarshal(taskBytes, apTask)
	if err != nil {
		return nil, err
	}
	return apTask, nil
}

// queryAndUpdateVIN processes message that: sends autopi command to get vin, polls webhook db for result, and sets user_devices. retries if needed. starts drivly task if able to get vin.
func (ats *autoPiTaskService) queryAndUpdateVIN(ctx context.Context, taskID, deviceID, unitID, userDeviceID string) error {
	log := ats.log.With().Str("handler", queryAndUpdateVINTask).
		Str("taskID", taskID).
		Str("userDeviceID", userDeviceID).
		Str("unitID", unitID).
		Str("deviceID", deviceID).Logger()
	// store the userID?
	log.Info().Msg("Started processing autopi update")

	err := ats.updateTaskState(taskID, "Started", InProcess, 110, nil)
	if err != nil {
		log.Err(err).Msg("failed to persist state to redis")
		return err
	}
	//send command to update device, retry after 1m if get an error
	cmd, err := ats.autoPiSvc.CommandQueryVIN(ctx, unitID, deviceID, userDeviceID)
	if err != nil {
		log.Err(err).Msg("failed to call autopi api query vin command")
		_ = ats.updateTaskState(taskID, "autopi api call failed", Failure, 500, err)
		return err
	}
	//check that command did not timeout
	backoffSchedule := []time.Duration{
		10 * time.Second,
		30 * time.Second,
		30 * time.Second,
	}
	for _, backoff := range backoffSchedule {
		time.Sleep(backoff)
		job, _, err := ats.autoPiSvc.GetCommandStatus(ctx, cmd.Jid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				_ = ats.updateTaskState(taskID, "autopi job was not found in db", Failure, 500, err)
				log.Err(err).Msg("autopi job not found in db")
				return err
			}
			continue // try again if error
		}
		if job.CommandState == "COMMAND_EXECUTED" {
			if job.Result != nil {
				if len(job.Result.Value) == 17 {
					// update user_devices
					userDevice, err := models.FindUserDevice(ctx, ats.dbs().Reader, userDeviceID)
					if err != nil {
						_ = ats.updateTaskState(taskID, "failed to get user_device by id "+userDeviceID, Failure, 500, err)
						log.Err(err).Msg("failed to get user_device by id")
						return nil // we return nil b/c do not want to retry task in this situation, failure could be something bigger DB related
					}
					userDevice.VinIdentifier = null.StringFrom(job.Result.Value)
					userDevice.VinConfirmed = true

					_, err = userDevice.Update(ctx, ats.dbs().Writer, boil.Infer())
					if err != nil {
						_ = ats.updateTaskState(taskID, "failed to update user_device with VIN", Failure, 500, err)
						log.Err(err).Msg("failed to update user_device with VIN")
						return nil // we return nil b/c do not want to retry task in this situation, failure could be something bigger DB related
					}
					log = log.With().Str("vin", job.Result.Value).Logger()
				}
			}
			// the job was completed, but we still may have not found a valid VIN
			_ = ats.updateTaskState(taskID, "autopi query vin returned by device", Success, 200, nil)
			break
		}
		if job.CommandState == "TIMEOUT" {
			_ = ats.updateTaskState(taskID, "autopi query vin job timed out, device may be offline or rebooting", Failure, 400, nil)
			log.Warn().Msg("autopi query vin job timed out")
			return errors.New("job timeout")
		}
	}
	log.Info().Msg("Succesfully queried device for vin. if no vin field in log, no fin was returned.")
	return nil // by not returning error, task will not be retried.
}

// updateTaskState updates the status of the task in redis
func (ats *autoPiTaskService) updateTaskState(taskID, description string, status TaskStatusEnum, code int, err error) error {
	updateCnt := 0
	existing, _ := ats.GetTaskStatus(context.Background(), taskID)
	if existing != nil {
		updateCnt = existing.Updates + 1
	}
	t := AutoPiTask{
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
	set := ats.redis.Set(context.Background(), buildAutoPiTaskRedisKey(taskID), jb, time.Hour*72)
	return set.Err()
}

func buildAutoPiTaskRedisKey(taskID string) string {
	return updateAutoPiTask + "_" + taskID
}

// AutoPiTask describes a task that is being worked on asynchronously for autopi
type AutoPiTask struct {
	TaskID      string    `json:"taskId"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Code        int       `json:"code"`
	Error       *string   `json:"error,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
	// Updates increments every time the job was updated.
	Updates int `json:"updates"`
}

type TaskStatusEnum string

const (
	Pending   TaskStatusEnum = "Pending"
	InProcess TaskStatusEnum = "InProcess"
	Success   TaskStatusEnum = "Success"
	Failure   TaskStatusEnum = "Failure"
)

// StandardRedis combines methods of redis client and what taskq expects so can use it for both clustered redis client and regular client
type StandardRedis interface {
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Pipelined(ctx context.Context, fn func(pipe redis.Pipeliner) error) ([]redis.Cmder, error)

	// Eval Required by redislock
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(ctx context.Context, scripts ...string) *redis.BoolSliceCmd
	ScriptLoad(ctx context.Context, script string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}
