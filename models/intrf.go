package models

type Mountser interface {
	Mounts() map[string][]string
}

type Userser interface {
	Users() []string
}
