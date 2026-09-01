package domain

import "time"

// Node описывает характеристики вычислительного узла.
type Node struct {
	Version   string
	Region    string
	Resources ResourcesMax
	StartedAt time.Time
}
