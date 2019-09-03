/*
Copyright 2016 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mqtrigger

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/fission/fission/pkg/crd"
	"github.com/fission/fission/pkg/mqtrigger/messageQueue"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Start(logger *zap.Logger, routerUrl string) error {
	fissionClient, kubeClient, _, err := crd.MakeFissionClient()
	if err != nil {
		return errors.Wrap(err, "failed to get fission or kubernetes client")
	}

	err = fissionClient.WaitForCRDs()
	if err != nil {
		return errors.Wrap(err, "error waiting for CRDs")
	}

	// Message queue type: nats is the only supported one for now
	mqType := os.Getenv("MESSAGE_QUEUE_TYPE")
	mqUrl := os.Getenv("MESSAGE_QUEUE_URL")

	// For authentication with message queue
	mqSecretName := os.Getenv("MESSAGE_QUEUE_SECRETS")

	var secrets *v1.Secret
	if mqSecretName != "" {
		secrets, _ = kubeClient.CoreV1().Secrets(getCurrentNamespace()).Get(mqSecretName, metav1.GetOptions{})
	}

	mqCfg := messageQueue.MessageQueueConfig{
		MQType:  mqType,
		Url:     mqUrl,
		Secrets: secrets,
	}
	messageQueue.MakeMessageQueueTriggerManager(logger, fissionClient, routerUrl, mqCfg)
	return nil
}

func getCurrentNamespace() string {
	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")

	namespace := strings.TrimSpace(string(data))
	if err == nil && len(namespace) > 0 {
		return namespace
	}

	return "default"
}
