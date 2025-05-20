// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package beater

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"runtime"

	"encoding/json"

	"github.com/logrhythm/pubsubbeat/lrutilities/crypto"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"cloud.google.com/go/pubsub"
	"github.com/logrhythm/pubsubbeat/config"
	"github.com/logrhythm/pubsubbeat/environment"
	"github.com/logrhythm/pubsubbeat/heartbeat"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Pubsubbeat struct {
	done         chan struct{}
	config       *config.Config
	client       beat.Client
	pubsubClient *pubsub.Client
	subscription *pubsub.Subscription
	logger       *logp.Logger
	StopChan     chan struct{}
}

const (
	cycleTime = 10 //will be in seconds
	// ServiceName is the name of the service
	ServiceName = "pubsubbeat"
)

var (
	receivedLogsInCycle int64
	counterLock         sync.RWMutex
	logsReceived        int64
	fqBeatName          string
)

var stopCh = make(chan struct{})

// New creates an instance of pubsubbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config, err := config.GetAndValidateConfig(cfg)
	if err != nil {
		return nil, err
	}

	logger := logp.NewLogger(fmt.Sprintf("PubSub: %s/%s/%s", config.Project, config.Topic, config.Subscription.Name))
	logger.Infof("config retrieved: %+v", config)

	client, err := createPubsubClient(config)
	if err != nil {
		return nil, err
	}

	subscription, err := getOrCreateSubscription(client, config)
	if err != nil {
		return nil, err
	}
	fqBeatName = os.Getenv(environment.FQBeatName)

	logp.Info("Config fields: %+v", config)
	logp.Info("Fully Qualified Beatname: %s", fqBeatName)

	bt := &Pubsubbeat{
		done:         make(chan struct{}),
		config:       config,
		pubsubClient: client,
		subscription: subscription,
		logger:       logger,
	}

	return bt, nil
}

// Run executes an instance of pubsubbeat.
func (bt *Pubsubbeat) Run(b *beat.Beat) error {

	bt.logger.Info("pubsubbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-bt.done
		// The beat is stopping...
		bt.logger.Info("cancelling PubSub receive context...")
		cancel()
		bt.logger.Info("closing PubSub client...")
		bt.pubsubClient.Close()
	}()

	// Self-reporting heartbeat
	bt.StopChan = make(chan struct{})
	hb := heartbeat.NewHeartbeatConfig(bt.config.HeartbeatInterval, bt.config.HeartbeatDisabled)
	heartbeater, err := hb.CreateEnabled(bt.StopChan, ServiceName)
	if err != nil {
		logp.Info("Error while creating new heartbeat object: %v", err)
	}
	if heartbeater != nil {
		heartbeater.Start(bt.StopChan, bt.client.Publish)
	}

	go cycleRoutine(time.Duration(cycleTime))

	err = bt.subscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		// This callback is invoked concurrently by multiple goroutines

		var datetime time.Time
		eventMap := common.MapStr{
			"type":                   b.Info.Name,
			"message_id":             m.ID,
			"publish_time":           m.PublishTime,
			"message":                string(m.Data),
			"fullyqualifiedbeatname": fqBeatName,
		}

		if len(m.Attributes) > 0 {
			eventMap["attributes"] = m.Attributes
		}

		if bt.config.Json.Enabled {
			var unmarshalErr error
			if bt.config.Json.FieldsUnderRoot {
				unmarshalErr = json.Unmarshal(m.Data, &eventMap)
				bt.logger.Info("pubsubbeat eventMap 160:-", eventMap)
				if unmarshalErr == nil && bt.config.Json.FieldsUseTimestamp {
					var timeErr error
					timestamp := eventMap[bt.config.Json.FieldsTimestampName]
					delete(eventMap, bt.config.Json.FieldsTimestampName)
					datetime, timeErr = time.Parse(bt.config.Json.FieldsTimestampFormat, timestamp.(string))
					if timeErr != nil {
						bt.logger.Errorf("Failed to format timestamp string as time. Using time.Now(): %s", timeErr)
					}
				}
			} else {
				var jsonData interface{}
				unmarshalErr = json.Unmarshal(m.Data, &jsonData)
				if unmarshalErr == nil {
					//bt.logger.Info("pubsubbeat eventMap 174:-", resource.labels.service)
					bt.logger.Info("pubsubbeat eventMap 175:-", jsonData)
					eventMap["json"] = jsonData
				}
			}

			if unmarshalErr != nil {
				bt.logger.Errorf("failed to decode json message: %s", unmarshalErr)
				if bt.config.Json.AddErrorKey {
					eventMap["error"] = common.MapStr{
						"key":     "json",
						"message": fmt.Sprintf("failed to decode json message: %s", unmarshalErr),
					}
				}
			}
		}

		if datetime.IsZero() {
			datetime = time.Now()
		}

		bt.client.Publish(beat.Event{
			Timestamp: datetime,
			Fields:    eventMap,
		})

		counterLock.Lock()
		receivedLogsInCycle = receivedLogsInCycle + 1
		counterLock.Unlock()

		// TODO: Evaluate using AckHandler.
		m.Ack()
	})

	if err != nil {
		return fmt.Errorf("fail to receive message from subscription %q: %v", bt.subscription.String(), err)
	}

	return nil
}

// Stop a running instance of pubsubbeat.
func (bt *Pubsubbeat) Stop() {
	bt.client.Close()
	close(stopCh)
	bt.StopChan <- struct{}{}
	close(bt.done)
}

func createPubsubClient(config *config.Config) (*pubsub.Client, error) {
	ctx := context.Background()
	userAgent := fmt.Sprintf(
		"Elastic/Pubsubbeat (%s; %s)", runtime.GOOS, runtime.GOARCH)
	tempFile, err := ioutil.TempFile(path.Dir(config.CredentialsFile), "temp")
	if err != nil {
		return nil, fmt.Errorf("fail to create temp file for decrypted credentials: %v", err)
	}
	defer os.Remove(tempFile.Name())

	options := []option.ClientOption{option.WithUserAgent(userAgent)}
	if config.CredentialsFile != "" {
		c, err := ioutil.ReadFile(config.CredentialsFile) // just pass the file name
		if err != nil {
			return nil, fmt.Errorf("fail to encrypted credentials: %v", err)
		}

		decryptedContent, err := crypto.Decrypt(string(c))
		if err != nil {
			return nil, errors.New("error decrypting Content")
		}
		tempFile.WriteString(decryptedContent)
		options = append(options, option.WithCredentialsFile(tempFile.Name()))

	}

	client, err := pubsub.NewClient(ctx, config.Project, options...)
	if err != nil {
		return nil, fmt.Errorf("fail to create pubsub client: %v", err)
	}
	return client, nil
}

func getOrCreateSubscription(client *pubsub.Client, config *config.Config) (*pubsub.Subscription, error) {
	if !config.Subscription.Create {
		subscription := client.Subscription(config.Subscription.Name)
		return subscription, nil
	}

	topic := client.Topic(config.Topic)
	ctx := context.Background()

	subscription, err := client.CreateSubscription(ctx, config.Subscription.Name, pubsub.SubscriptionConfig{
		Topic:               topic,
		RetainAckedMessages: config.Subscription.RetainAckedMessages,
		RetentionDuration:   config.Subscription.RetentionDuration,
	})

	st, ok := status.FromError(err)
	if ok && st.Code() == codes.AlreadyExists {
		// The subscription already exists.
		subscription = client.Subscription(config.Subscription.Name)
	} else if err != nil {
		return nil, fmt.Errorf(st.Message())
	} else if ok && st.Code() == codes.NotFound {
		return nil, fmt.Errorf("topic %q does not exists", config.Topic)
	} else if !ok {
		return nil, fmt.Errorf("fail to create subscription: %v", err)
	}

	return subscription, nil
}

func cycleRoutine(n time.Duration) {
	for {
		select {
		case <-stopCh:
			break
		default:
		}

		time.Sleep(n * time.Second)
		counterLock.Lock()
		logsReceived = logsReceived + receivedLogsInCycle
		var recordsPerSecond int64
		if receivedLogsInCycle > 0 {
			recordsPerSecond = receivedLogsInCycle / int64(cycleTime)
		}
		logp.Info("Total number of logs received in current cycle :  %d", receivedLogsInCycle)
		receivedLogsInCycle = 0
		counterLock.Unlock()
		logp.Info("Total number of logs received :  %d", logsReceived)
		logp.Info("Events Flush Rate:  %v messages per second", recordsPerSecond)
	}
}
