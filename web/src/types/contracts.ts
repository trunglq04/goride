import { Coordinate, Driver, Route, RouteFare, Trip } from "./models";

// These are the endpoints the API Gateway must have for the frontend to work correctly
export enum BackendEndpoints {
  PREVIEW_TRIP = "/trip/preview",
  START_TRIP = "/trip/start",
  CANCEL_TRIP = "/trip/cancel",
  WS_DRIVERS = "/drivers",
  WS_RIDERS = "/riders",
  AUTH_REGISTER = "/auth/register",
  AUTH_LOGIN = "/auth/login",
  AUTH_VERIFY_OTP = "/auth/verify-otp",
  AUTH_RESEND_OTP = "/auth/resend-otp",
  AUTH_REFRESH = "/auth/refresh",
  AUTH_LOGOUT = "/auth/logout",
  AUTH_ME = "/auth/me",
}

export enum TripEvents {
  NoDriversFound = "trip.event.no_drivers_found",
  DriverAssigned = "trip.event.driver_assigned",
  Completed = "trip.event.completed",
  Cancelled = "trip.event.cancelled",
  Created = "trip.event.created",
  DriverLocation = "driver.cmd.location",
  DriverTripRequest = "driver.cmd.trip_request",
  DriverTripAccept = "driver.cmd.trip_accept",
  DriverTripDecline = "driver.cmd.trip_decline",
  DriverTripCanceled = "driver.cmd.trip_canceled",
  DriverRegister = "driver.cmd.register",
  PaymentSessionCreated = "payment.event.session_created",
}

// Messages sent from the server to the client via the websocket
export type ServerWsMessage =
  | PaymentSessionCreatedRequest
  | DriverAssignedRequest
  | DriverLocationRequest
  | DriverTripRequest
  | DriverTripCanceledRequest
  | DriverRegisterRequest
  | TripCreatedRequest
  | NoDriversFoundRequest;

// Messages sent from the client to the server via the websocket
export type ClientWsMessage = DriverResponseToTripResponse | DriverLocationClientMessage;

export interface DriverLocationClientMessage {
  type: TripEvents.DriverLocation;
  data: {
    driverID: string;
    location: Coordinate;
    geohash: string;
  };
}

export interface TripCreatedRequest {
  type: TripEvents.Created;
  data: Trip;
}

export interface NoDriversFoundRequest {
  type: TripEvents.NoDriversFound;
}

export interface DriverRegisterRequest {
  type: TripEvents.DriverRegister;
  data: Driver;
}

export interface DriverTripRequest {
  type: TripEvents.DriverTripRequest;
  data: Trip;
}

export interface DriverTripCanceledRequest {
  type: TripEvents.DriverTripCanceled;
  data: Trip;
}

export interface PaymentEventSessionCreatedData {
  tripID: string;
  sessionID: string;
  amount: number;
  currency: string;
}

export interface PaymentSessionCreatedRequest {
  type: TripEvents.PaymentSessionCreated;
  data: PaymentEventSessionCreatedData;
}

export interface DriverAssignedRequest {
  type: TripEvents.DriverAssigned;
  data: Trip;
}

export interface DriverLocationRequest {
  type: TripEvents.DriverLocation;
  data: Driver[];
}

export interface DriverResponseToTripResponse {
  type: TripEvents.DriverTripAccept | TripEvents.DriverTripDecline;
  data: {
    tripID: string;
    riderID: string;
    driver: Driver;
  };
}

export interface HTTPTripPreviewResponse {
  route: Route;
  rideFares: RouteFare[];
}

export interface HTTPTripStartRequestPayload {
  rideFareID: string;
  userID: string;
}

export interface HTTPTripPreviewRequestPayload {
  userID: string;
  pickup: Coordinate;
  destination: Coordinate;
}

export function isValidTripEvent(event: string): event is TripEvents {
  return Object.values(TripEvents).includes(event as TripEvents);
}

export function isValidWsMessage(message: ServerWsMessage): message is ServerWsMessage {
  return isValidTripEvent(message.type);
}
