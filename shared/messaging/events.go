package messaging

import (
	pbd "github.com/trunglq04/goride/shared/proto/driver"
	pbt "github.com/trunglq04/goride/shared/proto/trip"
)

const (
	// Trip related queues
	FindAvailableDriversQueue       = "find_available_drivers"
	DriverCmdTripRequestQueue       = "driver_cmd_trip_request" // send req to driver if they accept the trip
	DriverTripResponseQueue         = "driver_trip_response"    // send res to rider if driver accepted or not
	NotifyDriverNoDriversFoundQueue = "notify_driver_no_drivers_found"
	NotifyDriverAssignQueue         = "notify_driver_assign"
	NotifyDriverLocationQueue       = "notify_driver_location"
	// Payment related queues
	PaymentTripResponseQueue         = "payment_trip_response"
	NotifyPaymentSessionCreatedQueue = "notify_payment_session_created"
	NotifyPaymentSuccessQueue        = "payment_success"
	DeadLetterQueue                  = "dead_letter_queue"
)

type TripEventData struct {
	Trip             *pbt.Trip `json:"trip"`
	ExcludeDriverIDs []string  `json:"excludedDriverIDs,omitempty"`
}

type DriverTripResponseData struct {
	Driver  *pbd.Driver `json:"driver"`
	TripID  string      `json:"tripID"`
	RiderID string      `json:"riderID"`
}

type PaymentEventSessionCreatedData struct {
	TripID    string  `json:"tripID"`
	SessionID string  `json:"sessionID"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

type PaymentTripResponseData struct {
	TripID   string  `json:"tripID"`
	UserID   string  `json:"userID"`
	DriverID string  `json:"driverID"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type PaymentStatusUpdateData struct {
	TripID   string `json:"tripID"`
	UserID   string `json:"userID"`
	DriverID string `json:"driverID"`
}
