package gatewayapp

import (
	"context"
	"testing"

	"aegis/dto"
	chaossystemmodule "aegis/module/chaossystem"
	labelmodule "aegis/module/label"
)

type resourceLabelClientStub struct {
	enabled bool
}

func (s *resourceLabelClientStub) Enabled() bool { return s.enabled }

func (s *resourceLabelClientStub) CreateLabel(context.Context, *labelmodule.CreateLabelReq) (*labelmodule.LabelResp, error) {
	return &labelmodule.LabelResp{ID: 3, Key: "env", Value: "prod"}, nil
}

func (s *resourceLabelClientStub) GetLabel(context.Context, int) (*labelmodule.LabelDetailResp, error) {
	return &labelmodule.LabelDetailResp{LabelResp: labelmodule.LabelResp{ID: 3, Key: "env", Value: "prod"}}, nil
}

func (s *resourceLabelClientStub) ListLabels(context.Context, *labelmodule.ListLabelReq) (*dto.ListResp[labelmodule.LabelResp], error) {
	return &dto.ListResp[labelmodule.LabelResp]{Items: []labelmodule.LabelResp{{ID: 3, Key: "env", Value: "prod"}}}, nil
}

func (s *resourceLabelClientStub) UpdateLabel(context.Context, *labelmodule.UpdateLabelReq, int) (*labelmodule.LabelResp, error) {
	return &labelmodule.LabelResp{ID: 3, Key: "env", Value: "prod"}, nil
}

func (s *resourceLabelClientStub) DeleteLabel(context.Context, int) error { return nil }

func (s *resourceLabelClientStub) BatchDeleteLabels(context.Context, []int) error { return nil }

func TestRemoteAwareLabelServiceRequiresResource(t *testing.T) {
	service := remoteAwareLabelService{}
	if _, err := service.List(context.Background(), &labelmodule.ListLabelReq{}); err == nil {
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

func (s *resourceChaosSystemClientStub) ListChaosSystems(context.Context, *chaossystemmodule.ListChaosSystemReq) (*dto.ListResp[chaossystemmodule.ChaosSystemResp], error) {
	return &dto.ListResp[chaossystemmodule.ChaosSystemResp]{Items: []chaossystemmodule.ChaosSystemResp{{ID: 8, Name: "k8s"}}}, nil
}

func (s *resourceChaosSystemClientStub) GetChaosSystem(context.Context, int) (*chaossystemmodule.ChaosSystemResp, error) {
	return &chaossystemmodule.ChaosSystemResp{ID: 8, Name: "k8s"}, nil
}

func (s *resourceChaosSystemClientStub) CreateChaosSystem(context.Context, *chaossystemmodule.CreateChaosSystemReq) (*chaossystemmodule.ChaosSystemResp, error) {
	return &chaossystemmodule.ChaosSystemResp{ID: 8, Name: "k8s"}, nil
}

func (s *resourceChaosSystemClientStub) UpdateChaosSystem(context.Context, *chaossystemmodule.UpdateChaosSystemReq, int) (*chaossystemmodule.ChaosSystemResp, error) {
	return &chaossystemmodule.ChaosSystemResp{ID: 8, Name: "k8s"}, nil
}

func (s *resourceChaosSystemClientStub) DeleteChaosSystem(context.Context, int) error { return nil }

func (s *resourceChaosSystemClientStub) UpsertChaosSystemMetadata(context.Context, int, *chaossystemmodule.BulkUpsertSystemMetadataReq) error {
	return nil
}

func (s *resourceChaosSystemClientStub) ListChaosSystemMetadata(context.Context, int, string) ([]chaossystemmodule.SystemMetadataResp, error) {
	return []chaossystemmodule.SystemMetadataResp{{ID: 1, SystemName: "k8s"}}, nil
}

func TestRemoteAwareChaosSystemServiceRequiresResource(t *testing.T) {
	service := remoteAwareChaosSystemService{}
	if _, err := service.ListSystems(context.Background(), &chaossystemmodule.ListChaosSystemReq{}); err == nil {
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
