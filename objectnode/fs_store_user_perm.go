package objectnode

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	updateUserPermInterval = time.Second * 5
	userPermLoaderNum      = 1
)

type Permissionx int

var (
	UserPermission Permissionx = 1
)

type userPermission struct {
	UserId     string      `json:"user_id"`
	Permission Permissionx `json:"permission"`
}

type UserPermissionStore interface {
	hasPermission(userId string) bool
	scheduleUpdate()
}

type userPermissionStore struct {
}

func NewUserPermissionStore() UserPermissionStore {
	return &userPermissionStore{}
}

func (s *userPermissionStore) fetchUserPermission() error {
	var data []userPermission
	resp, err := http.Get("http://10.237.96.202:3003/api/acl/list")
	if err != nil {
		log.Println(err)
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		log.Println(err)
		return err
	}
	log.Printf("data %+v", data)
	return nil
}

func (s *userPermissionStore) hasPermission(userId string) bool {
	return false
}

func (s *userPermissionStore) scheduleUpdate() {
	ticker := time.NewTicker(updateUserPermInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.fetchUserPermission()
		}
	}
}
