//go:build !ignore_autogenerated

/*
SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and cap-operator-lifecycle contributors
SPDX-License-Identifier: Apache-2.0
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPOperator) DeepCopyInto(out *CAPOperator) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPOperator.
func (in *CAPOperator) DeepCopy() *CAPOperator {
	if in == nil {
		return nil
	}
	out := new(CAPOperator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CAPOperator) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPOperatorList) DeepCopyInto(out *CAPOperatorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CAPOperator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPOperatorList.
func (in *CAPOperatorList) DeepCopy() *CAPOperatorList {
	if in == nil {
		return nil
	}
	out := new(CAPOperatorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CAPOperatorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPOperatorSpec) DeepCopyInto(out *CAPOperatorSpec) {
	*out = *in
	in.SubscriptionServer.DeepCopyInto(&out.SubscriptionServer)
	if in.IngressGatewayLabels != nil {
		in, out := &in.IngressGatewayLabels, &out.IngressGatewayLabels
		*out = make([]NameValue, len(*in))
		copy(*out, *in)
	}
	in.Controller.DeepCopyInto(&out.Controller)
	in.Monitoring.DeepCopyInto(&out.Monitoring)
	in.Webhook.DeepCopyInto(&out.Webhook)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPOperatorSpec.
func (in *CAPOperatorSpec) DeepCopy() *CAPOperatorSpec {
	if in == nil {
		return nil
	}
	out := new(CAPOperatorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPOperatorStatus) DeepCopyInto(out *CAPOperatorStatus) {
	*out = *in
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPOperatorStatus.
func (in *CAPOperatorStatus) DeepCopy() *CAPOperatorStatus {
	if in == nil {
		return nil
	}
	out := new(CAPOperatorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertManager) DeepCopyInto(out *CertManager) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertManager.
func (in *CertManager) DeepCopy() *CertManager {
	if in == nil {
		return nil
	}
	out := new(CertManager)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificateConfig) DeepCopyInto(out *CertificateConfig) {
	*out = *in
	out.Gardener = in.Gardener
	out.CertManager = in.CertManager
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificateConfig.
func (in *CertificateConfig) DeepCopy() *CertificateConfig {
	if in == nil {
		return nil
	}
	out := new(CertificateConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Controller) DeepCopyInto(out *Controller) {
	*out = *in
	if in.VersionMonitoring != nil {
		in, out := &in.VersionMonitoring, &out.VersionMonitoring
		*out = new(VersionMonitoring)
		**out = **in
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]v1.Volume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.VolumeMounts != nil {
		in, out := &in.VolumeMounts, &out.VolumeMounts
		*out = make([]v1.VolumeMount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Controller.
func (in *Controller) DeepCopy() *Controller {
	if in == nil {
		return nil
	}
	out := new(Controller)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Gardener) DeepCopyInto(out *Gardener) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Gardener.
func (in *Gardener) DeepCopy() *Gardener {
	if in == nil {
		return nil
	}
	out := new(Gardener)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Grafana) DeepCopyInto(out *Grafana) {
	*out = *in
	if in.Dashboard != nil {
		in, out := &in.Dashboard, &out.Dashboard
		*out = new(GrafanaDashboard)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Grafana.
func (in *Grafana) DeepCopy() *Grafana {
	if in == nil {
		return nil
	}
	out := new(Grafana)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GrafanaDashboard) DeepCopyInto(out *GrafanaDashboard) {
	*out = *in
	if in.ConfigMapLabels != nil {
		in, out := &in.ConfigMapLabels, &out.ConfigMapLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GrafanaDashboard.
func (in *GrafanaDashboard) DeepCopy() *GrafanaDashboard {
	if in == nil {
		return nil
	}
	out := new(GrafanaDashboard)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Monitoring) DeepCopyInto(out *Monitoring) {
	*out = *in
	if in.Grafana != nil {
		in, out := &in.Grafana, &out.Grafana
		*out = new(Grafana)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Monitoring.
func (in *Monitoring) DeepCopy() *Monitoring {
	if in == nil {
		return nil
	}
	out := new(Monitoring)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NameValue) DeepCopyInto(out *NameValue) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NameValue.
func (in *NameValue) DeepCopy() *NameValue {
	if in == nil {
		return nil
	}
	out := new(NameValue)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubscriptionServer) DeepCopyInto(out *SubscriptionServer) {
	*out = *in
	if in.CertificateConfig != nil {
		in, out := &in.CertificateConfig, &out.CertificateConfig
		*out = new(CertificateConfig)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubscriptionServer.
func (in *SubscriptionServer) DeepCopy() *SubscriptionServer {
	if in == nil {
		return nil
	}
	out := new(SubscriptionServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionMonitoring) DeepCopyInto(out *VersionMonitoring) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionMonitoring.
func (in *VersionMonitoring) DeepCopy() *VersionMonitoring {
	if in == nil {
		return nil
	}
	out := new(VersionMonitoring)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Webhook) DeepCopyInto(out *Webhook) {
	*out = *in
	if in.CertificateConfig != nil {
		in, out := &in.CertificateConfig, &out.CertificateConfig
		*out = new(CertificateConfig)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Webhook.
func (in *Webhook) DeepCopy() *Webhook {
	if in == nil {
		return nil
	}
	out := new(Webhook)
	in.DeepCopyInto(out)
	return out
}
