package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/iskorotkov/bully-election-dashboard/pkg/state"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Collector struct {
	namespace string
	timeout   time.Duration
	clientset *kubernetes.Clientset
	logger    *zap.Logger
}

func NewCollector(namespace string, timeout time.Duration, logger *zap.Logger) (*Collector, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("couldn't create kubernetes config",
			zap.Error(err))
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error("couldn't create kubernetes clientset",
			zap.Error(err))
		return nil, err
	}

	return &Collector{
		namespace: namespace,
		timeout:   timeout,
		clientset: clientset,
		logger:    logger,
	}, nil
}

func (c *Collector) Collect() ([]state.ReplicaState, error) {
	logger := c.logger.Named("collect")
	logger.Debug("collection started")

	pods, err := c.pods(c.namespace, c.timeout, logger.Named("pods"))
	if err != nil {
		logger.Error("couldn't list pods",
			zap.String("namespace", c.namespace),
			zap.Error(err))
		return nil, err
	}

	logger.Debug("all pods info was fetched",
		zap.Any("pods", pods))

	results := collectFromPods(pods, c.timeout, logger.Named("collect-from-pods"))

	data := make([]state.ReplicaState, 0)
	for state := range results {
		data = append(data, state)
	}

	logger.Debug("all data was fetched",
		zap.Any("data", data))

	return data, nil
}

func (c *Collector) pods(namespace string, timeout time.Duration, logger *zap.Logger) (*corev1.PodList, error) {
	logger.Debug("fetching list of pods",
		zap.String("namespace", namespace))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pods, nil
}

func collectFromPods(pods *corev1.PodList, timeout time.Duration, logger *zap.Logger) <-chan state.ReplicaState {
	results := make(chan state.ReplicaState, len(pods.Items))
	defer close(results)

	wg := sync.WaitGroup{}
	wg.Add(len(pods.Items))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, pod := range pods.Items {
		pod := pod

		go func() {
			defer wg.Done()

			logger := logger.Named(pod.GetName())

			// Push unknown state if and only if failed to fetch metrics from pod.
			success := false
			defer func() {
				if !success {
					results <- state.NewUnknownReplicaState(pod.GetName())
				}
			}()

			if pod.Status.PodIP == "" {
				return
			}

			url := fmt.Sprintf("http://%s/metrics", pod.Status.PodIP)
			req, err := http.NewRequestWithContext(ctx, "get", url, nil)
			if err != nil {
				logger.Error("couldn't create request",
					zap.String("url", url),
					zap.Error(err))
				return
			}

			logger.Debug("preparing request for pod",
				zap.String("pod", pod.GetName()),
				zap.String("url", url),
				zap.Any("request", req))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				logger.Error("couldn't execute request",
					zap.String("url", url),
					zap.Any("request", req),
					zap.Error(err))
				return
			}

			logger.Debug("got response from pod",
				zap.String("pod", pod.GetName()),
				zap.String("url", url),
				zap.Any("response", resp))

			defer func() {
				if err := resp.Body.Close(); err != nil {
					logger.Error("response body close failed",
						zap.Error(err))
				}
			}()

			if resp.StatusCode != http.StatusOK {
				logger.Error("response returned incorrect status code",
					zap.Int("code", resp.StatusCode))
				return
			}

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error("couldn't read response body",
					zap.Any("response", resp),
					zap.Error(err))
				return
			}

			logger.Debug("data from pod",
				zap.String("pod", pod.GetName()),
				zap.String("data", string(b)))

			var data state.ReplicaState
			if err := json.Unmarshal(b, &data); err != nil {
				logger.Error("couldn't unmarshal response body",
					zap.Any("response", resp),
					zap.Error(err))
				return
			}

			results <- data

			// Mark metrics fetch as succeeded to avoid pushing unknown state.
			success = true
		}()
	}

	wg.Wait()

	return results
}
