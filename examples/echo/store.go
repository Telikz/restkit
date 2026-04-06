package main

import (
	"errors"
	"sync"
	"sync/atomic"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserStore struct {
	mu     sync.RWMutex
	users  map[int64]User
	nextID int64
}

func NewUserStore() *UserStore {
	return &UserStore{
		users: map[int64]User{
			1: {ID: 1, Name: "Alice", Email: "alice@example.com"},
			2: {ID: 2, Name: "Bob", Email: "bob@example.com"},
		},
		nextID: 3,
	}
}

func (s *UserStore) Create(name, email string) User {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := atomic.AddInt64(&s.nextID, 1) - 1
	user := User{ID: int(id), Name: name, Email: email}
	s.users[id] = user
	return user
}

func (s *UserStore) Get(id int64) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	if !ok {
		return User{}, errors.New("user not found")
	}
	return user, nil
}

func (s *UserStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]User, 0, len(s.users))
	for _, user := range s.users {
		list = append(list, user)
	}
	return list
}
