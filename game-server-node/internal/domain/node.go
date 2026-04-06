package domain

import "time"

type Node struct {
	Version   string
	Region    string
	Resources ResourcesMax
	StartedAt time.Time
}
