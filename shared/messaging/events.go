package messaging

import pb "github.com/trunglq04/goride/shared/proto/trip"

const (
	FindAvailbleDriverQueue = "find_available_drivers"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}
