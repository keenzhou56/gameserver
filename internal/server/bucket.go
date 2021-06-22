package server

import (
	"errors"
	"fmt"
	"gameserver/internal/server/conf"
	"sync"
	"sync/atomic"
)

// Bucket ...
type Bucket struct {
	c *conf.Bucket
	//sync.RWMutex
	//mapUser   map[int]*User
	mapUser   sync.Map
	Online    int32
	MaxOnLine int32
}

// InitBucket ...
func (m *Bucket) InitBucket() {
	//if m.mapUser == nil {
	//	m.mapUser = make(map[int]*User)
	//}
}

// NewBucket new a bucket struct. store the key with im channel.
func NewBucket(c *conf.Bucket) (b *Bucket) {
	b = new(Bucket)
	b.c = c
	b.InitBucket()
	return
}

// UnsafeGetUser ...
func (m *Bucket) UnsafeGetUser(userID int64) (*User, error) {
	if userID <= 0 {
		errMsg := fmt.Sprintf("Error: Bucket.UnsafeGetUser: User.UserID must larger than 0, given: %d", userID)
		return nil, errors.New(errMsg)
	}
	//user, existsFlag := m.mapUser[userID]
	user, existsFlag := m.mapUser.Load(userID)
	if !existsFlag {
		errMsg := fmt.Sprintf("Error: Bucket.UnsafeGetUser: User not found, UserID: %d", userID)
		return nil, errors.New(errMsg)
	}
	return user.(*User), nil
}

// GetUser ...
func (m *Bucket) GetUser(userID int64) (*User, error) {
	//m.RLock()
	//defer m.RUnlock()
	user, err := m.UnsafeGetUser(userID)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return user, nil
}

// UnsafeSetUser ...
func (m *Bucket) UnsafeSetUser(user *User) error {
	if user.UserID <= 0 {
		return errors.New("Bucket.UnsafeSetUser: User.UserId must larger than 0")
	}
	//m.mapUser[user.UserID] = user
	m.mapUser.Store(user.UserID, user)
	return nil
}

// AddUser ...
func (m *Bucket) AddUser(user *User) error {
	if user.UserID <= 0 {
		return errors.New("Bucket.AddUser :User.UserId must larger than 0")
	}
	//m.Lock()
	//m.mapUser[user.UserID] = user
	//m.Online++
	m.mapUser.Store(user.UserID, user)
	newCnt := atomic.AddInt32(&m.Online, 1)
	if newCnt > m.MaxOnLine {
		//todo 释放当前user
		m.MaxOnLine = m.Online
	}
	//m.Unlock()
	return nil
}

// SetUser ...
func (m *Bucket) SetUser(user *User) error {
	if user.UserID <= 0 {
		return errors.New("Bucket.SetUser :User.UserId must larger than 0")
	}
	//m.Lock()
	//m.mapUser[user.UserID] = user
	//m.Unlock()
	m.mapUser.Store(user.UserID, user)
	return nil
}

// TestAndSet ...
func (m *Bucket) TestAndSet(user *User) interface{} {
	//m.Lock()
	//defer m.Unlock()
	//if v, ok := m.mapUser[user.UserID]; ok {
	//	return v
	//}
	//m.mapUser[user.UserID] = user
	if v, ok := m.mapUser.Load(user.UserID); ok {
		return v
	}
	m.mapUser.Store(user.UserID, user)
	return nil
}

// UnsafeDelUser ...
func (m *Bucket) UnsafeDelUser(userID int64) error {
	if userID <= 0 {
		return errors.New("Bucket.UnsafeDelUser: User.userID must larger than 0")
	}
	//_, existsFlag := m.mapUser[userID]
	//if !existsFlag {
	//	return errors.New("Bucket.UnsafeDelUser: User not found")
	//}
	//delete(m.mapUser, userID)
	//m.Online--
	_, existsFlag := m.mapUser.Load(userID)
	if !existsFlag {
		return errors.New("Bucket.UnsafeDelUser: User not found")
	}
	m.mapUser.Delete(userID)
	atomic.AddInt32(&m.Online, -1)
	return nil
}

// DelUser ...
func (m *Bucket) DelUser(userID int64) error {
	if userID <= 0 {
		return errors.New("Bucket.DelUser: User.userID must larger than 0")
	}
	//m.Lock()
	//_, existsFlag := m.mapUser[userID]
	//if !existsFlag {
	//	m.Unlock()
	//	return errors.New("Bucket.DelUser: User not found")
	//}
	//delete(m.mapUser, userID)
	//m.Online--
	//m.Unlock()
	_, existsFlag := m.mapUser.LoadAndDelete(userID)
	if !existsFlag {
		return errors.New("Bucket.DelUser: User not found")
	}
	atomic.AddInt32(&m.Online, -1)

	return nil
}

// UnsafeLenUser ...
func (m *Bucket) UnsafeLenUser() int {
	//if m.mapUser == nil {
	//	return 0
	//}
	//return len(m.mapUser)
	return int(m.Online)

}

// LenUser ...
func (m *Bucket) LenUser() int {
	//m.RLock()
	//defer m.RUnlock()
	return m.UnsafeLenUser()
}

// UnsafeRangeUser ...
func (m *Bucket) UnsafeRangeUser(f func(user *User)) {
	//	if m.mapUser == nil {
	//		return
	//	}
	//	for _, user := range m.mapUser {
	//		f(user)
	//	}
}
func (m *Bucket) walkUser(key, value interface{}) bool {
	// user回调方法
	// Usage: m.mapUser.Range(m.walkUser)
	return true
}

// RLockRangeUser ...
func (m *Bucket) RLockRangeUser(f func(user *User)) {
	//m.RLock()
	//defer m.RUnlock()
	//m.UnsafeRangeUser(f)
	m.mapUser.Range(func(key, value interface{}) bool {
		f(value.(*User))
		return true
	})
}

// LockRangeUser ...
func (m *Bucket) LockRangeUser(f func(user *User)) {
	//m.Lock()
	//defer m.Unlock()
	m.UnsafeRangeUser(f)
}

// ExistsUser ...
func (m *Bucket) ExistsUser(userID int64) bool {
	//if userID <= 0 {
	//	return false
	//}
	//m.RLock()
	//_, existsFlag := m.mapUser[userID]
	//m.RUnlock()
	_, existsFlag := m.mapUser.Load(userID)
	return existsFlag
}

// GetMapUser ...
// todo 使用sync.map.range做深度copy
func (m *Bucket) GetMapUser() map[int64]*User {
	//return m.mapUser
	return make(map[int64]*User, 0)
}

// DelUserGroupID ...
func (m *Bucket) DelUserGroupID(userID int64, groupID string) error {
	if userID <= 0 {
		return errors.New("Bucket.DelUserGroupID: User.userID must larger than 0")
	}
	//m.Lock()
	//user, existsFlag := m.mapUser[userID]
	//if !existsFlag {
	//	m.Unlock()
	//	return errors.New("Bucket.DelUserGroupID: User not found")
	//}
	//delete(user.GroupIDs, groupID)
	//m.mapUser[userID] = user
	//m.Unlock()
	user, existsFlag := m.mapUser.Load(userID)
	if !existsFlag {
		return errors.New("Bucket.DelUserGroupID: User not found")
	}
	delete(user.(*User).GroupIDs, groupID)
	m.mapUser.Store(userID, user)

	return nil
}

// JoinUserGroupID ...
func (m *Bucket) JoinUserGroupID(userID int64, groupID string) error {
	if userID <= 0 {
		return errors.New("Bucket.DelUserGroupID: User.userID must larger than 0")
	}
	//m.Lock()
	//user, existsFlag := m.mapUser[userID]
	//if !existsFlag {
	//	m.Unlock()
	//	return errors.New("Bucket.JoinUserGroupID: User not found")
	//}
	//user.GroupIDs[groupID] = groupID
	//m.mapUser[userID] = user
	//m.Unlock()
	user, existsFlag := m.mapUser.Load(userID)
	if !existsFlag {
		return errors.New("Bucket.JoinUserGroupID: User not found")
	}
	user.(*User).GroupIDs[groupID] = groupID
	m.mapUser.Store(userID, user)
	return nil
}
