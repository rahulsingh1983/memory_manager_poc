package memory

type Handle struct {
	id uint64
}

type PlacementStrategy string

const (
	PlacementFirstFit PlacementStrategy = "first-fit"
)

type Config struct {
	DiskSize          int
	PlacementStrategy PlacementStrategy
}

type Stats struct {
	TotalBytes    int
	UsedBytes     int
	FreeBytes     int
	ActiveHandles int
}