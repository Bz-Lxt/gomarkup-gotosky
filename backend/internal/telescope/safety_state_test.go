package telescope_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gotosky/gotosky/internal/domain"
	"github.com/gotosky/gotosky/internal/telescope"
)

type rainOnConnectDriver struct {
	*telescope.MockDriver
}

func (*rainOnConnectDriver) Connect(context.Context) error {
	return &telescope.DeviceError{Code: "RainDetected", Class: telescope.ClassSafety, Msg: "rain"}
}

func (*rainOnConnectDriver) Heartbeat(context.Context) error { return nil }
func (*rainOnConnectDriver) Park(context.Context) error      { return nil }

func TestSafetyFaultFinishesInErrorAfterParking(t *testing.T) {
	driver := &rainOnConnectDriver{MockDriver: telescope.NewMockDriver()}
	session := domain.Session{
		ID:         uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		RigID:      uuid.MustParse("20000000-0000-4000-8000-000000000001"),
		State:      "IDLE",
		SourceMode: "SIMULATED",
	}
	actor := telescope.NewActor(session, driver, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	actor.Start(ctx)
	defer actor.Stop()

	commandID := uuid.MustParse("30000000-0000-4000-8000-000000000001")
	_ = actor.Submit(commandID, "START", nil)
	got := actor.Snapshot()

	if got.State != "ERROR" {
		t.Errorf("state = %q, want ERROR", got.State)
	}
	if !strings.Contains(got.LastError, "RainDetected") {
		t.Errorf("last error = %q, want RainDetected", got.LastError)
	}
	if got.EndedAt == nil {
		t.Error("ended_at is nil after safety shutdown")
	}
}
