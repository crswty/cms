package datastore

import (
	"crswty.com/cms/server"
	"fmt"
	"sort"
)

type Memory struct {
	Data map[string]map[string]server.Object
}

type Record struct {
	Id   string
	Type server.Type
	Data server.Object
}

func NewMemory(records ...Record) (Memory, error) {
	memory := Memory{Data: map[string]map[string]server.Object{}}

	for _, record := range records {
		err := memory.Create(record.Type, record.Id, record.Data)
		if err != nil {
			return Memory{}, fmt.Errorf("error creating memory store with inital data id: %s error: %w", record.Id, err)
		}
	}

	return memory, nil
}

func (m Memory) List(t server.Type) ([]server.Object, error) {
	keys := make([]string, 0)
	for k, _ := range m.Data[t.Name] {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	objs := make([]server.Object, 0)
	for _, key := range keys {
		objs = append(objs, m.Data[t.Name][key])
	}

	return objs, nil
}

func (m Memory) Get(t server.Type, id string) (server.Object, error) {
	allOfType, typeFound := m.Data[t.Name]
	if !typeFound {
		return server.Object{}, fmt.Errorf("no type with name %s found in storage", t.Name)
	}
	obj, objectFound := allOfType[id]
	if !objectFound {
		return server.Object{}, fmt.Errorf("no object with id %s found in storage", id)
	}
	return obj, nil
}

func (m Memory) Create(t server.Type, id string, obj server.Object) error {
	_, typeFound := m.Data[t.Name]
	if !typeFound {
		m.Data[t.Name] = map[string]server.Object{}
	}
	m.Data[t.Name][id] = obj
	return nil
}

func (m Memory) Update(t server.Type, id string, obj server.Object) error {
	return m.Create(t, id, obj)
}

func (m Memory) Delete(t server.Type, id string) error {
	delete(m.Data[t.Name], id)
	return nil
}
