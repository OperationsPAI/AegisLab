package gateway

import (
	"context"
	"testing"

	"aegis/dto"
	chaossystem "aegis/module/chaossystem"
	label "aegis/module/label"
)

type resourceLabelClientStub struct {
	enabled bool
}

func (s *resourceLabelClientStub) Enabled() bool { return s.enabled }

func (s *resourceLabelClientStub) CreateLabel(context.Context, *label.CreateLabelReq) (*label.LabelResp, error) {
	return &label.LabelResp{ID: 3, Key: "env", Value: "prod"}, nil
}

func (s *resourceLabelClientStub) GetLabel(context.Context, int) (*label.LabelDetailResp, error) {
	return &label.LabelDetailResp{LabelResp: label.LabelResp{ID: 3, Key: "env", Value: "prod"}}, nil
}

func (s *resourceLabelClientStub) ListLabels(context.Context, *label.ListLabelReq) (*dto.ListResp[label.LabelResp], error) {
	return &dto.ListResp[label.LabelResp]{Items: []label.LabelResp{{ID: 3, Key: "env", Value: "prod"}}}, nil
}

func (s *resourceLabelClientStub) UpdateLabel(context.Context, *label.UpdateLabelReq, int) (*label.LabelResp, error) {
	return &label.LabelResp{ID: 3, Key: "env", Value: "prod"}, nil
}

func (s *resourceLabelClientStub) DeleteLabel(context.Context, int) error { return nil }

func (s *resourceLabelClientStub) BatchDeleteLabels(context.Context, []int) error { return nil }

func TestRemoteAwareLabelServiceRequiresResource(t *testing.T) {
	service := remoteAwareLabelService{}
	if _, err := service.List(context.Background(), &label.ListLabelReq{}); err == nil {
		t.Fatal("List() error = nil, want missing dependency")
	}
}

func TestRemoteAwareLabelServiceUsesResourceClient(t *testing.T) {
	service := remoteAwareLabelService{resource: &resourceLabelClientStub{enabled: true}}
	resp, err := service.GetDetail(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetDetail() error = %v", err)
	}
	if resp.ID != 3 || resp.Key != "env" {
		t.Fatalf("GetDetail() unexpected response: %+v", resp)
	}
}

type resourceChaosSystemClientStub struct {
	enabled bool
}

func (s *resourceChaosSystemClientStub) Enabled() bool { return s.enabled }

func (s *resourceChaosSystemClientStub) ListChaosSystems(context.Context, *chaossystem.ListChaosSystemReq) (*dto.ListResp[chaossystem.ChaosSystemResp], error) {
	return &dto.ListResp[chaossystem.ChaosSystemResp]{Items: []chaossystem.ChaosSystemResp{{ID: 8, Name: "k8s"}}}, nil
}

func (s *resourceChaosSystemClientStub) GetChaosSystem(context.Context, int) (*chaossystem.ChaosSystemResp, error) {
	return &chaossystem.ChaosSystemResp{ID: 8, Name: "k8s"}, nil
}

func (s *resourceChaosSystemClientStub) CreateChaosSystem(context.Context, *chaossystem.CreateChaosSystemReq) (*chaossystem.ChaosSystemResp, error) {
	return &chaossystem.ChaosSystemResp{ID: 8, Name: "k8s"}, nil
}

func (s *resourceChaosSystemClientStub) UpdateChaosSystem(context.Context, *chaossystem.UpdateChaosSystemReq, int) (*chaossystem.ChaosSystemResp, error) {
	return &chaossystem.ChaosSystemResp{ID: 8, Name: "k8s"}, nil
}

func (s *resourceChaosSystemClientStub) DeleteChaosSystem(context.Context, int) error { return nil }

func (s *resourceChaosSystemClientStub) UpsertChaosSystemMetadata(context.Context, int, *chaossystem.BulkUpsertSystemMetadataReq) error {
	return nil
}

func (s *resourceChaosSystemClientStub) ListChaosSystemMetadata(context.Context, int, string) ([]chaossystem.SystemMetadataResp, error) {
	return []chaossystem.SystemMetadataResp{{ID: 1, SystemName: "k8s"}}, nil
}

func TestRemoteAwareChaosSystemServiceRequiresResource(t *testing.T) {
	service := remoteAwareChaosSystemService{}
	if _, err := service.ListSystems(context.Background(), &chaossystem.ListChaosSystemReq{}); err == nil {
		t.Fatal("ListSystems() error = nil, want missing dependency")
	}
}

func TestRemoteAwareChaosSystemServiceUsesResourceClient(t *testing.T) {
	service := remoteAwareChaosSystemService{resource: &resourceChaosSystemClientStub{enabled: true}}
	resp, err := service.GetSystem(context.Background(), 8)
	if err != nil {
		t.Fatalf("GetSystem() error = %v", err)
	}
	if resp.ID != 8 || resp.Name != "k8s" {
		t.Fatalf("GetSystem() unexpected response: %+v", resp)
	}
}
