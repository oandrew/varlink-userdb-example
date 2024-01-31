package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/varlink/go/varlink"
)

func getMachineID() (string, error) {
	data, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(data)), nil

}

type svc struct {
	machineID string
}

func (s *svc) GetUserRecord(ctx context.Context, c VarlinkCall, uid_ *int64, userName_ *string, service_ string) error {
	type StatusInfo struct {
		Service string `json:"service"`
	}
	type UserRecord struct {
		Username string                `json:"userName"`
		UID      int64                 `json:"uid"`
		GID      int64                 `json:"gid"`
		HomeDir  string                `json:"homeDirectory"`
		Shell    string                `json:"shell"`
		Status   map[string]StatusInfo `json:"status"`
	}
	if userName_ != nil && *userName_ == "foobar" || uid_ != nil && *uid_ == 30117 {
		j, err := json.Marshal(UserRecord{
			Username: "foobar",
			UID:      30117,
			GID:      30117,
			HomeDir:  "/nonexistant",
			Shell:    "/sbin/nologin",
			Status: map[string]StatusInfo{
				s.machineID: StatusInfo{Service: "test"},
			},
		})
		if err != nil {
			panic(err)
		}
		return c.ReplyGetUserRecord(ctx, json.RawMessage(j), false)
	} else {
		return c.ReplyNoRecordFound(ctx)
	}

}
func (s *svc) GetGroupRecord(ctx context.Context, c VarlinkCall, gid_ *int64, groupName_ *string, service_ string) error {
	return c.ReplyMethodNotImplemented(ctx, "GetGroupRecord")
}
func (s *svc) GetMemberships(ctx context.Context, c VarlinkCall, userName_ *string, groupName_ *string, service_ string) error {
	return c.ReplyMethodNotImplemented(ctx, "GetMemberships")
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background())
	machineID, err := getMachineID()
	if err != nil {
		log.Fatal(err)
	}

	service, _ := varlink.NewService(
		"Example",
		"This",
		"1",
		"https://example.org/this",
	)

	svc := svc{machineID: machineID}
	service.RegisterInterface(VarlinkNew(&svc))
	path := "/run/systemd/userdb/test"
	if err := service.Bind(ctx, fmt.Sprintf("unix:%s", path)); err != nil {
		panic(err)
	}
	os.Chmod(path, 0777)
	go func() {
		log.Fatal(service.DoListen(ctx, 0))
	}()
	<-ctx.Done()
	service.Shutdown()

}
