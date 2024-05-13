package objectnode

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	UserPermissionType     = 1
	updateUserPermInterval = time.Second * 5
	userPermLoaderNum      = 1
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
	req *http.Request
}

func NewUserPermissionStore() (UserPermissionStore, error) {
	req, err := http.NewRequest(http.MethodGet, `http://10.237.96.202:3003/api/acl/list?type=1`, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	store := &userPermissionStore{
		req,
	}
	fmt.Printf("userPermission %+v", store)
	go store.scheduleUpdate()

	return store, nil
}

func (s *userPermissionStore) fetchUserPermission() error {
	var data []userPermission
	resp, err := http.DefaultClient.Do(s.req)
	if err != nil {
		log.Println("error 1", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("req %+v \n", s.req)
		log.Printf("resp %+v \n", resp)
		return errors.New("sync user permission failed")
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Println("error 2", err)
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
			fmt.Println("start sync")
			if err := s.fetchUserPermission(); err != nil {
				log.Println("err", err)
			}
		}
	}
}
