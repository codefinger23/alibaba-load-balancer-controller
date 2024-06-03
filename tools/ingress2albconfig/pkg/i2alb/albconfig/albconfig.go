package albconfig

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type AlbConfig struct {
	metav1.TypeMeta `json:",inline" yaml:",inline" `
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"  protobuf:"bytes,1,opt,name=metadata"`

	// Spec is the desired state of the Gateway.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec AlbConfigSpec `json:"spec,omitempty" yaml:"spec,omitempty"  protobuf:"bytes,2,opt,name=spec"`
}

type AlbConfigSpec struct {
	LoadBalancer *LoadBalancerSpec `json:"config,omitempty" yaml:"config,omitempty"  protobuf:"bytes,1,rep,name=config"`
	Listeners    []*ListenerSpec   `json:"listeners,omitempty" yaml:"listeners,omitempty"  protobuf:"bytes,2,rep,name=listeners"`
}

type ListenerSpec struct {
	GzipEnabled    *bool              `json:"gzipEnabled,omitempty" yaml:"gzipEnabled,omitempty"  protobuf:"bytes,1,opt,name=gzipEnabled"`
	Port           intstr.IntOrString `json:"port,omitempty" yaml:"port,omitempty"  protobuf:"bytes,5,opt,name=port"`
	Protocol       string             `json:"protocol,omitempty" yaml:"protocol,omitempty"  protobuf:"bytes,8,opt,name=protocol"`
	IdleTimeout    int                `json:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty"  protobuf:"bytes,10,opt,name=idleTimeout"`
	Description    string             `json:"description,omitempty" yaml:"description,omitempty"  protobuf:"bytes,13,opt,name=description"`
	CaEnabled      bool               `json:"caEnabled,omitempty" yaml:"caEnabled,omitempty"  protobuf:"bytes,14,opt,name=caEnabled"`
	RequestTimeout int                `json:"requestTimeout,omitempty" yaml:"requestTimeout,omitempty"  protobuf:"bytes,16,opt,name=requestTimeout"`
}

// LoadBalancer is a nested struct in alb response
type LoadBalancerSpec struct {
	Id                        string          `json:"id,omitempty" yaml:"id,omitempty" protobuf:"bytes,1,opt,name=id"`
	Name                      string          `json:"name,omitempty" yaml:"name,omitempty" protobuf:"bytes,2,opt,name=name"`
	AddressAllocatedMode      string          `json:"addressAllocatedMode,omitempty" yaml:"addressAllocatedMode,omitempty" protobuf:"bytes,3,opt,name=addressAllocatedMode"`
	AddressType               string          `json:"addressType,omitempty" yaml:"addressType,omitempty" protobuf:"bytes,4,opt,name=addressType"`
	Ipv6AddressType           string          `json:"ipv6AddressType,omitempty" yaml:"ipv6AddressType,omitempty" protobuf:"bytes,5,opt,name=ipv6AddressType"`
	AddressIpVersion          string          `json:"addressIpVersion,omitempty" yaml:"addressIpVersion,omitempty" protobuf:"bytes,6,opt,name=addressIpVersion"`
	ResourceGroupId           string          `json:"resourceGroupId,omitempty" yaml:"resourceGroupId,omitempty" protobuf:"bytes,7,opt,name=resourceGroupId"`
	Edition                   string          `json:"edition,omitempty" yaml:"edition,omitempty" protobuf:"bytes,8,opt,name=edition"`
	ZoneMappings              []ZoneMapping   `json:"zoneMappings,omitempty" yaml:"zoneMappings,omitempty" protobuf:"bytes,9,rep,name=zoneMappings"`
	AccessLogConfig           AccessLogConfig `json:"accessLogConfig,omitempty" yaml:"accessLogConfig,omitempty" protobuf:"bytes,10,opt,name=accessLogConfig"`
	DeletionProtectionEnabled *bool           `json:"deletionProtectionEnabled,omitempty" yaml:"deletionProtectionEnabled,omitempty" protobuf:"bytes,11,opt,name=deletionProtectionEnabled"`
	ForceOverride             *bool           `json:"forceOverride,omitempty" yaml:"forceOverride,omitempty" protobuf:"bytes,13,opt,name=forceOverride"`
	Tags                      []Tag           `json:"tags,omitempty" yaml:"tags,omitempty" protobuf:"bytes,15,opt,name=tags"`
	ListenerForceOverride     *bool           `json:"listenerForceOverride,omitempty" yaml:"listenerForceOverride,omitempty" protobuf:"bytes,16,opt,name=listenerForceOverride"`
}

type Tag struct {
	Key   string `json:"key,omitempty" yaml:"key,omitempty"  protobuf:"bytes,1,opt,name=key"`
	Value string `json:"value,omitempty" yaml:"value,omitempty"  protobuf:"bytes,2,opt,name=value"`
}

type ZoneMapping struct {
	VSwitchId string `json:"vSwitchId,omitempty" yaml:"vSwitchId,omitempty"  protobuf:"bytes,1,opt,name=vSwitchId"`
}

type AccessLogConfig struct {
	LogStore   string `json:"logStore,omitempty" yaml:"logStore,omitempty"  protobuf:"bytes,1,opt,name=logStore"`
	LogProject string `json:"logProject,omitempty" yaml:"logProject,omitempty"  protobuf:"bytes,2,opt,name=logProject"`
}

func (c *AlbConfig) GetObjectKind() schema.ObjectKind {
	return &c.TypeMeta
}

func (c *AlbConfig) DeepCopyObject() runtime.Object {
	if c == nil {
		return nil
	}
	out := new(AlbConfig)
	*out = *c
	return out
}
