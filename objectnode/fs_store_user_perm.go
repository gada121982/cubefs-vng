package objectnode

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/cubefs/cubefs/util/config"
)

type Permissionx int

var userNormalPermission Permissionx = 1

const (
	UserPermissionType     = 1
	updateUserPermInterval = time.Second * 5

	configAclServiceEndpoint = "aclServiceEndpoint"
	configAclServiceToken    = "aclServiceToken"
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

func NewUserPermissionStore(cfg *config.Config) (UserPermissionStore, error) {
	endpoint := cfg.GetString(configAclServiceEndpoint)
	token := cfg.GetString(configAclServiceToken)
	if endpoint == "" || token == "" {
		return nil, errors.New("missing acl service configuration")
	}
	req, err := http.NewRequest(http.MethodGet, endpoint+`/api/acl/list?type=1`, nil)
	if err != nil {
		return nil, err
	}
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
