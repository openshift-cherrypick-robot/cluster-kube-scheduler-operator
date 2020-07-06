package configmetrics

import (
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/component-base/metrics/legacyregistry"

	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	configlisters "github.com/openshift/client-go/config/listers/config/v1"
)

func Register(configInformer configinformers.SharedInformerFactory) {
	legacyregistry.MustRegister(&configMetrics{
		configLister: configInformer.Config().V1().Schedulers().Lister(),
		config: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cluster_master_schedulable",
			Help: "Reports whether the cluster master nodes are schedulable.",
		}),
		policy: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cluster_legacy_scheduler_policy",
			Help: "Reports whether the cluster scheduler is configured with a legacy KubeScheduler Policy.",
		}),
		nodeSelector: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cluster_default_node_selector",
			Help: "Reports whether the cluster scheduler is configured with a default node selector.",
		}),
	})
}

// configMetrics implements metrics gathering for this component.
type configMetrics struct {
	configLister configlisters.SchedulerLister
	config       prometheus.Gauge
	policy       prometheus.Gauge
	nodeSelector prometheus.Gauge
}

func (m *configMetrics) ClearState() {}

func (m *configMetrics) Create(version *semver.Version) bool {
	return true
}

// Describe reports the metadata for metrics to the prometheus collector.
func (m *configMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.config.Desc()
}

// Collect calculates metrics from the cached config and reports them to the prometheus collector.
func (m *configMetrics) Collect(ch chan<- prometheus.Metric) {
	if config, err := m.configLister.Get("cluster"); err == nil {
		ch <- booleanGaugeValue(m.config, config.Spec.MastersSchedulable)
		ch <- booleanGaugeValue(m.policy, len(config.Spec.Policy.Name) > 0)
		ch <- booleanGaugeValue(m.nodeSelector, len(config.Spec.DefaultNodeSelector) > 0)
	}
}

func (m *configMetrics) FQName() string {
	return "cluster_kube_scheduler_operator"
}

func booleanGaugeValue(g prometheus.Gauge, value bool) prometheus.Gauge {
	if value {
		g.Set(1)
	} else {
		g.Set(0)
	}
	return g
}
