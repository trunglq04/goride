package contracts

// AmqpMessage is the message structure for AMQP.
type AmqpMessage struct {
	OwnerID string `json:"ownerId"`
	Data    []byte `json:"data"`
}

// Routing keys - using consistent event/command patterns
const (
	// Trip events (trip.event.*)
	TripEventCreated             = "trip.event.created"
	TripEventCanceled            = "trip.event.canceled"
	TripEventNoDriversFound      = "trip.event.no_drivers_found"
	TripEventDriverAssigned      = "trip.event.driver_assigned"
	TripEventDriverNotInterested = "trip.event.driver_not_interested"

	// Driver commands (driver.cmd.*)
	DriverCmdTripRequest  = "driver.cmd.trip_request"
	DriverCmdTripAccept   = "driver.cmd.trip_accept"
	DriverCmdTripDecline  = "driver.cmd.trip_decline"
	DriverCmdTripCanceled = "driver.cmd.trip_canceled"
	DriverCmdLocation     = "driver.cmd.location"
	DriverCmdRegister     = "driver.cmd.register"

	// Payment events (payment.event.*)
	PaymentEventSessionCreated = "payment.event.session_created"
	PaymentEventSuccess        = "payment.event.success"
	PaymentEventFailed         = "payment.event.Failed"
	PaymentEventCancelled      = "payment.event.cancelled"

	// Payment commands (payment.cmd.*)
	PaymentCmdCreateSession = "payment.cmd.create_session"
)
