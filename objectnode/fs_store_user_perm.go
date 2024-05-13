package objectnode

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

type Permissionx int

var userNormalPermission Permissionx = 1

const (
	UserPermissionType     = 1
	updateUserPermInterval = time.Second * 5
	token                  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI2NWVlYjljMWE3Mzg5OTgzNWZjNDk1YzUiLCJuYW1lIjoiaGFpbnY0IiwiaXNMb2ciOmZhbHNlLCJyb2xlIjoiQWRtaW4iLCJpYXQiOjE3MTM3NjY2NjksImV4cCI6MTcxMzg1MzA2OX0.eSO-um58kj3gCJI8Iz_CzZSU18-QuDMh9O6w8C6H1V0"
)

type userPermission struct {
	UserId     string `json:"user_id"`
	Permission int    `json:"permission"`
}

type UserPermissionStore interface {
	hasPermission(userId string) bool
	scheduleUpdate()
}

type userPermissionStore struct {
	req        *http.Request
	permission sync.Map
}

// TODO: get from config file
func NewUserPermissionStore() (UserPermissionStore, error) {
	req, err := http.NewRequest(http.MethodGet, `http://10.237.96.202:3003/api/acl/list?type=1`, nil)
	if err != nil {
		return nil, err
	}
	// TODO: get token from config file
	req.Header.Add("Authorization", "Bearer "+token)
	store := new(userPermissionStore)
	store.req = req
	go store.scheduleUpdate()

	return store, nil
}

func (s *userPermissionStore) fetchUserPermission() error {
	var data []userPermission
	resp, err := http.DefaultClient.Do(s.req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("sync user permission failed")
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	for _, val := range data {
		s.permission.Store(val.UserId, Permissionx(val.Permission))
	}
	return nil
}

func (s *userPermissionStore) hasPermission(userId string) bool {
	p, ok := s.permission.Load(userId)
	log.Println("hasPermission", p, ok)
	if !ok {
		return false
	}
	if p == userNormalPermission {
		return false
	}
	return true
}

func (s *userPermissionStore) scheduleUpdate() {
	ticker := time.NewTicker(updateUserPermInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := s.fetchUserPermission(); err != nil {
				log.Println("err", err) // TODO: add logging library here
			}
		}
	}
}
