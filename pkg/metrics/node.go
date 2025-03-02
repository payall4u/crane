package metrics

import (
	"github.com/gocrane/crane/pkg/known"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
)

const (
	CraneNodeSubsystem = "node"
	CranePodSubsystem  = "pod"
)

var (
	podElasticCPUDesc = prometheus.NewDesc("crane_pod_elastic_cpu_request",
		"The elastic cpu requested by pod",
		[]string{"pod", "namespace"},
		nil,
	)
	podElasticMemoryDesc = prometheus.NewDesc("crane_pod_elastic_memory_request",
		"The elastic cpu requested by pod",
		[]string{"pod", "namespace"},
		nil,
	)
)

func NewPodResourceCollector(podLister v1.PodLister) *PodResourceCollector {
	return &PodResourceCollector{
		podLister: podLister,
	}
}

type PodResourceCollector struct {
	podLister v1.PodLister
}

func (n *PodResourceCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- podElasticCPUDesc
	descs <- podElasticMemoryDesc
}

func (n *PodResourceCollector) Collect(metrics chan<- prometheus.Metric) {
	pods, err := n.podLister.List(labels.Everything())
	if err != nil {
		klog.ErrorS(err, "list pods failed")
		return
	}

	for _, pod := range pods {
		eCPU, eMemory := resource.NewQuantity(0, resource.DecimalSI), resource.NewQuantity(0, resource.BinarySI)
		for _, container := range pod.Spec.Containers {
			eCPU.Add(*container.Resources.Requests.Name(known.ElasticCPU, resource.DecimalSI))
			eMemory.Add(*container.Resources.Requests.Name(known.ElasticCPU, resource.DecimalSI))
		}
		if eCPU.IsZero() && eMemory.IsZero() {
			continue
		}
		metrics <- prometheus.MustNewConstMetric(podElasticCPUDesc, prometheus.GaugeValue, eCPU.AsApproximateFloat64(), pod.Name, pod.Namespace)
		metrics <- prometheus.MustNewConstMetric(podElasticMemoryDesc, prometheus.GaugeValue, eMemory.AsApproximateFloat64(), pod.Name, pod.Namespace)
	}
}

var (
	nodeElasticCPUDesc = prometheus.NewDesc("crane_node_elastic_cpu_allocatable",
		"The elastic cpu of the node.",
		[]string{"node"},
		nil,
	)
	nodeElasticMemoryDesc = prometheus.NewDesc("crane_node_elastic_memory_allocatable",
		"The elastic memory requested by pod",
		[]string{"node"},
		nil,
	)
	nodeCPUAllocatableDesc = prometheus.NewDesc("crane_node_cpu_allocatable",
		"The cpu allocatable of the node.",
		[]string{"node"},
		nil,
	)
	nodeCPUCapacityDesc = prometheus.NewDesc("crane_node_cpu_capacity",
		"The cpu capacity of the node.",
		[]string{"node"},
		nil,
	)
	nodeMemoryAllocatableDesc = prometheus.NewDesc("crane_node_memory_allocatable",
		"The memory allocatable requested by pod",
		[]string{"node"},
		nil,
	)
	nodeMemoryCapacityDesc = prometheus.NewDesc("crane_node_memory_capacity",
		"The memory capacity requested by pod",
		[]string{"node"},
		nil,
	)
	nodeCPUReservedDesc = prometheus.NewDesc("crane_node_cpu_reserved",
		"The reserved cpu of node",
		[]string{"node"},
		nil)
	nodeCPUUsageOnlineDesc = prometheus.NewDesc("crane_node_cpu_usage_online",
		"The online cpu usage of node",
		[]string{"node"},
		nil)
	nodeCPUUsageOfflineDesc = prometheus.NewDesc("crane_node_cpu_usage_offline",
		"The offline cpu usage of node",
		[]string{"node"},
		nil)
	nodeMemoryReservedDesc = prometheus.NewDesc("crane_node_memory_reserved",
		"The reserved memory of node",
		[]string{"node"},
		nil)
	nodeMemoryUsageOnlineDesc = prometheus.NewDesc("crane_node_memory_usage_online",
		"The online memory usage of node",
		[]string{"node"},
		nil)
	nodeMemoryUsageOfflineDesc = prometheus.NewDesc("crane_node_memory_usage_offline",
		"The offline memory usage of node",
		[]string{"node"},
		nil)
)

type NodeResourceCollector struct {
	nodeName           string
	nodeLister         v1.NodeLister
	nodeResourceGetter func() *known.ResourceStatus
}

func NewNodeResourceCollector(nodeName string, nodeLister v1.NodeLister, nodeResourceGetter func() *known.ResourceStatus) *NodeResourceCollector {
	return &NodeResourceCollector{
		nodeName:           nodeName,
		nodeLister:         nodeLister,
		nodeResourceGetter: nodeResourceGetter,
	}
}

func (n *NodeResourceCollector) Describe(descs chan<- *prometheus.Desc) {
	// resource metrics from status of node
	descs <- nodeElasticCPUDesc
	descs <- nodeElasticMemoryDesc
	descs <- nodeCPUAllocatableDesc
	descs <- nodeCPUCapacityDesc
	descs <- nodeMemoryAllocatableDesc
	descs <- nodeMemoryCapacityDesc

	// usage metrics
	descs <- nodeCPUReservedDesc
	descs <- nodeCPUUsageOnlineDesc
	descs <- nodeCPUUsageOfflineDesc
	descs <- nodeMemoryReservedDesc
	descs <- nodeMemoryUsageOnlineDesc
	descs <- nodeMemoryUsageOfflineDesc
}

func (n *NodeResourceCollector) Collect(metrics chan<- prometheus.Metric) {
	node, err := n.nodeLister.Get(n.nodeName)
	if err != nil {
		klog.ErrorS(err, "list pods failed")
		return
	}
	metrics <- prometheus.MustNewConstMetric(nodeElasticCPUDesc, prometheus.GaugeValue, node.Status.Allocatable.Name(known.ElasticCPU, resource.DecimalSI).AsApproximateFloat64(), node.Name)
	metrics <- prometheus.MustNewConstMetric(nodeElasticMemoryDesc, prometheus.GaugeValue, node.Status.Allocatable.Name(known.ElasticMemory, resource.BinarySI).AsApproximateFloat64(), node.Name)
	metrics <- prometheus.MustNewConstMetric(nodeCPUAllocatableDesc, prometheus.GaugeValue, node.Status.Allocatable.Cpu().AsApproximateFloat64(), node.Name)
	metrics <- prometheus.MustNewConstMetric(nodeMemoryAllocatableDesc, prometheus.GaugeValue, node.Status.Allocatable.Memory().AsApproximateFloat64(), node.Name)
	metrics <- prometheus.MustNewConstMetric(nodeCPUCapacityDesc, prometheus.GaugeValue, node.Status.Capacity.Cpu().AsApproximateFloat64(), node.Name)
	metrics <- prometheus.MustNewConstMetric(nodeMemoryCapacityDesc, prometheus.GaugeValue, node.Status.Capacity.Memory().AsApproximateFloat64(), node.Name)

	resourceStatus := n.nodeResourceGetter()
	if resourceStatus == nil {
		return
	}
	metrics <- prometheus.MustNewConstMetric(nodeCPUReservedDesc, prometheus.GaugeValue, resourceStatus.CPUReserved.AsApproximateFloat64(), node.Name)
	// TODO incorrect online define !!
	metrics <- prometheus.MustNewConstMetric(nodeCPUUsageOnlineDesc, prometheus.GaugeValue, resourceStatus.CPUUsage.AsApproximateFloat64(), node.Name)
	metrics <- prometheus.MustNewConstMetric(nodeCPUUsageOfflineDesc, prometheus.GaugeValue, resourceStatus.CPUUsageOffline.AsApproximateFloat64(), node.Name)

	metrics <- prometheus.MustNewConstMetric(nodeMemoryReservedDesc, prometheus.GaugeValue, resourceStatus.MemoryReserved.AsApproximateFloat64(), node.Name)
	// TODO incorrect online define !!
	metrics <- prometheus.MustNewConstMetric(nodeMemoryUsageOnlineDesc, prometheus.GaugeValue, resourceStatus.MemoryUsage.AsApproximateFloat64(), node.Name)
	metrics <- prometheus.MustNewConstMetric(nodeMemoryUsageOfflineDesc, prometheus.GaugeValue, resourceStatus.MemoryUsageOffline.AsApproximateFloat64(), node.Name)

}
