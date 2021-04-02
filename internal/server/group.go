package server

import (
	"errors"
	"fmt"
	"gameserver/pkg/config"
	"gameserver/pkg/protocal"
	"strconv"
	"sync"
)

// Groups ...
type Groups struct {
	sync.RWMutex
	mapGroup  map[string]*Group
	Online    int32
	MaxOnLine int32
}

// Group 房间信息
type Group struct {
	sync.RWMutex
	GroupID string
	UserIDs map[int64]int64
	Mq      chan *protocal.ImPacket
}

// NewGroup 生成新用户
func NewGroup() *Group {
	group := &Group{}
	group.UserIDs = make(map[int64]int64)
	return group
}

// NewGroups ...
func NewGroups() (g *Groups) {
	g = new(Groups)
	g.InitGroups()
	return
}

// InitGroups ...
func (m *Groups) InitGroups() {
	if m.mapGroup == nil {
		m.mapGroup = make(map[string]*Group)
	}
}

// CheckGroupIDValid 判断队伍id是否合法
func CheckGroupIDValid(groupID string) error {
	if len(groupID) < config.GroupIDLengthMin {
		return errors.New("group.CheckgroupIDValid groupID length must >= " + strconv.Itoa(config.GroupIDLengthMin))
	}

	if len(groupID) > config.GroupIDLengthMax {
		return errors.New("group.CheckgroupIDValid groupID length must <= " + strconv.Itoa(config.GroupIDLengthMax))
	}

	// TODO 判断首字母是否是英文字母

	return nil
}

// UnsafeGet ...
func (m *Groups) UnsafeGet(GroupID string) (*Group, error) {

	group, existsFlag := m.mapGroup[GroupID]
	if !existsFlag {
		errMsg := fmt.Sprintf("Groups.UnsafeGet: not found group, GroupID: %s", GroupID)
		return nil, errors.New(errMsg)
	}
	return group, nil
}

// Get ...
func (m *Groups) Get(GroupID string) (*Group, error) {
	if err := CheckGroupIDValid(GroupID); err != nil {
		return nil, err
	}
	m.RLock()
	group, existsFlag := m.mapGroup[GroupID]
	if !existsFlag {
		m.RUnlock()
		errMsg := fmt.Sprintf("Groups.Get: not found group, GroupID: %s", GroupID)
		return nil, errors.New(errMsg)
	}
	m.RUnlock()
	return group, nil
}

// UnsafeSet ...
func (m *Groups) UnsafeSet(group *Group) error {
	m.mapGroup[group.GroupID] = group
	return nil
}

// Add ...
func (m *Groups) Add(group *Group) error {
	if err := CheckGroupIDValid(group.GroupID); err != nil {
		return err
	}
	m.Lock()
	m.mapGroup[group.GroupID] = group
	m.Online++
	if m.Online > m.MaxOnLine {
		m.MaxOnLine = m.Online
	}
	m.Unlock()
	return nil
}

// Set ...
func (m *Groups) Set(group *Group) error {
	if err := CheckGroupIDValid(group.GroupID); err != nil {
		return err
	}
	m.Lock()
	m.mapGroup[group.GroupID] = group
	m.Unlock()
	return nil
}

// TestAndSet ...
func (m *Groups) TestAndSet(group *Group) interface{} {
	m.Lock()
	defer m.Unlock()
	if v, ok := m.mapGroup[group.GroupID]; ok {
		return v
	}
	m.mapGroup[group.GroupID] = group
	return nil

}

// UnsafeDel ...
func (m *Groups) UnsafeDel(GroupID string) error {
	_, existsFlag := m.mapGroup[GroupID]
	if !existsFlag {
		errMsg := fmt.Sprintf("Groups.UnsafeDel: not found group, GroupID: %s", GroupID)
		return errors.New(errMsg)
	}
	delete(m.mapGroup, GroupID)
	m.Online--
	return nil
}

// Del ...
func (m *Groups) Del(GroupID string) error {
	if err := CheckGroupIDValid(GroupID); err != nil {
		return err
	}
	m.Lock()
	_, existsFlag := m.mapGroup[GroupID]
	if !existsFlag {
		m.Unlock()
		errMsg := fmt.Sprintf("Groups.Del: not found group, GroupID: %s", GroupID)
		return errors.New(errMsg)
	}
	delete(m.mapGroup, GroupID)
	m.Online--
	m.Unlock()
	return nil
}

// UnsafeLen ...
func (m *Groups) UnsafeLen() int {
	if m.mapGroup == nil {
		return 0
	}
	return len(m.mapGroup)
}

// Len ...
func (m *Groups) Len() int {
	m.RLock()
	defer m.RUnlock()
	return m.UnsafeLen()
}

// UnsafeRange ...
func (m *Groups) UnsafeRange(f func(group *Group)) {
	if m.mapGroup == nil {
		return
	}
	for _, group := range m.mapGroup {
		f(group)
	}
}

// RLockRange ...
func (m *Groups) RLockRange(f func(group *Group)) {
	m.RLock()
	defer m.RUnlock()
	m.UnsafeRange(f)
}

// LockRange ...
func (m *Groups) LockRange(f func(group *Group)) {
	m.Lock()
	defer m.Unlock()
	m.UnsafeRange(f)
}

// Exists ...
func (m *Groups) Exists(GroupID string) bool {
	m.RLock()
	_, existsFlag := m.mapGroup[GroupID]
	m.RUnlock()
	return existsFlag
}

// UnsafeExists ...
func (m *Groups) UnsafeExists(GroupID string) bool {
	_, existsFlag := m.mapGroup[GroupID]
	return existsFlag
}

// GetData ...
func (m *Groups) GetData() map[string]*Group {
	return m.mapGroup
}

// BatchDelUser 移除组用户
func (m *Groups) BatchDelUser(groupIDs map[string]string, userID int64) error {
	m.Lock()
	for _, groupID := range groupIDs {
		group, existsFlag := m.mapGroup[groupID]
		if !existsFlag {
			continue
		}
		// 若组成员只剩下一个了，直接删除组
		if len(group.UserIDs) <= 1 {
			delete(m.mapGroup, groupID)
		} else {
			delete(group.UserIDs, userID)
			m.mapGroup[groupID] = group
		}
	}
	m.Unlock()
	return nil
}

// DelUser 移除组用户
func (m *Groups) DelUser(groupID string, userID int64) error {
	if len(groupID) <= 0 {
		return errors.New("Groups.DelUser: groupID len must larger than 0")
	}
	if userID <= 0 {
		return errors.New("Groups.DelUser: userID must larger than 0")
	}
	m.Lock()
	group, existsFlag := m.mapGroup[groupID]
	if !existsFlag {
		return errors.New("Groups.DelUser: group not found")
	}

	// 若组成员只剩下一个了，直接删除组
	if len(group.UserIDs) <= 1 {
		delete(m.mapGroup, groupID)
	} else {
		delete(group.UserIDs, userID)
		m.mapGroup[groupID] = group
	}
	m.Unlock()
	return nil
}

// JoinGroup ...
func (m *Groups) JoinGroup(groupID string, userID int64) error {
	if len(groupID) <= 0 {
		return errors.New("Groups.JoinGroup: groupID len must larger than 0")
	}
	if userID <= 0 {
		return errors.New("Groups.JoinGroup: userID must larger than 0")
	}

	m.Lock()
	group, existsFlag := m.mapGroup[groupID]
	if !existsFlag {
		group = NewGroup()
		m.Online++
		if m.Online > m.MaxOnLine {
			m.MaxOnLine = m.Online
		}
	} else {
		group = m.mapGroup[groupID]
	}
	group.UserIDs[userID] = userID
	m.mapGroup[groupID] = group
	m.Unlock()
	return nil
}

// DelGroupUserID ...
func (m *Groups) DelGroupUserID(groupID string, userID int64) error {
	if len(groupID) <= 0 {
		return errors.New("Groups.DelGroupUserID: groupID len must larger than 0")
	}

	if userID <= 0 {
		return errors.New("Groups.DelGroupUserID: userID must larger than 0")
	}
	m.Lock()
	group, existsFlag := m.mapGroup[groupID]
	if !existsFlag {
		m.Unlock()
		errMsg := fmt.Sprintf("Groups.UnsafeGet: not found group, GroupID: %s", groupID)
		return errors.New(errMsg)
	}
	delete(group.UserIDs, userID)
	m.mapGroup[groupID] = group
	m.Unlock()
	return nil
}

// GetOnline ...
func (m *Groups) GetOnline() int {
	return len(m.mapGroup)
}

// GetMapGroup ...
func (m *Groups) GetMapGroup() map[string]*Group {
	return m.mapGroup
}
